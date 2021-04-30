package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/hhistogram/processor"
	"github.com/afoninsky-go/hhistogram/swagger"
	"github.com/afoninsky-go/logger"
	"github.com/prometheus/common/log"
)

type Config struct {
	// resulting histogram name
	HistogramName string
	// duration of one histogram
	HistogramSliceDuration time.Duration
	// http receiver of generated metrics
	OutputEndpoint string
	// folder with openapi schemas
	SpecFolder string
}

type Service struct {
	log *logger.Logger
	api *swagger.Swagger
	cfg Config

	processedCounter uint32
	totalCounter     uint32
	notFoundCounter  uint32
}

func NewHistogramService(cfg Config) (*Service, error) {
	s := &Service{}
	s.log = logger.NewSTDLogger()
	s.api = swagger.NewSwaggerRouter()
	s.cfg = cfg

	go func() {
		for {
			time.Sleep(time.Second * 10)
			if s.totalCounter > 0 {
				var notFoundPercent uint32
				if s.processedCounter > 0 {
					notFoundPercent = s.notFoundCounter * 100 / s.processedCounter
				} else {
					notFoundPercent = 100
				}
				log.Infof("Processed %d of %d events, %d%% not found in swager spec", s.processedCounter, s.totalCounter, notFoundPercent)
				s.totalCounter = 0
				s.notFoundCounter = 0
				s.processedCounter = 0
			}
		}
	}()

	return s, s.api.LoadSpecFolder(cfg.SpecFolder)
}

// convert incoming bulks of metrics into histograms
func (s *Service) BulkHandler(w http.ResponseWriter, r *http.Request) {
	// create bulk processor to convert stream of http events into set of histograms
	config := processor.NewConfig().WithName(s.cfg.HistogramName).WithSliceDuration(s.cfg.HistogramSliceDuration)
	bulk := processor.NewHistogramProcessor(config).WithInterceptor(s).WithLogger(s.log)

	// process incoming metrics
	if err := bulk.ReadFromStream(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get remote endpoint with passed query args
	u, _ := url.Parse(s.cfg.OutputEndpoint)
	u.RawQuery = r.URL.RawQuery

	// proxy processor response to endpoint
	buf := new(bytes.Buffer)
	bulk.Process(buf)
	proxyReq, _ := http.NewRequest(http.MethodPost, u.String(), buf)
	defer proxyReq.Body.Close()
	httpClient := &http.Client{
		Timeout: time.Second * 60,
	}
	proxyRes, err := httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer proxyRes.Body.Close()

	// send back response status code
	w.WriteHeader(proxyRes.StatusCode)
	bodyBytes, _ := ioutil.ReadAll(proxyRes.Body)
	w.Write(bodyBytes)
}

func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	// dummy health check
}

// convert http event to obfuscated metric
func (s *Service) OnMetric(m *metric.Metric) error {
	s.totalCounter++

	// ensure "url" and "method" labels exist in the source metric -> metric is valid
	rawurl, ok := m.Labels["url"]
	if !ok {
		s.log.Warn("metric doesn't have url label, ignoring ...")
		return nil
	}
	method, ok := m.Labels["method"]
	if !ok {
		s.log.Warn("metric doesn't have method label, ignoring ...")
		return nil
	}

	// parse url and remove high-cardinality label
	u, err := url.Parse(rawurl)
	if err != nil {
		s.log.Warnf("invalid url in metric %s, removing...", rawurl)
		m.SetName("")
		return nil
	}
	delete(m.Labels, "url")

	// add http-specific labels
	m.Labels["service"] = ""
	m.Labels["action"] = ""

	// search openapi schemas for specified url
	req := http.Request{
		Method: method,
		URL:    u,
	}
	route := s.api.TestRoute(&req)

	// route not found
	if route == nil {
		s.notFoundCounter++
		return nil
	}

	// route found
	m.Labels["service"] = route.Tag
	m.Labels["action"] = route.Path
	s.processedCounter++

	return nil
}
