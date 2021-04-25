package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/hhistogram/openapi"
	"github.com/afoninsky-go/hhistogram/processor"
	"github.com/afoninsky-go/logger"
)

const histogramName = "test"
const endpoint = "http://localhost:8081"

type Service struct {
	log *logger.Logger
	api *openapi.OpenAPI
}

func NewHistogramService() *Service {
	s := &Service{}
	s.log = logger.NewSTDLogger()
	s.api = openapi.NewURLParser().WithLogger(s.log)
	return s
}

// convert incoming bulks of metrics into histograms
func (s *Service) BulkHandler(w http.ResponseWriter, r *http.Request) {
	// create bulk processor to convert stream of http events into set of histograms
	config := processor.NewConfig().WithName(histogramName)
	bulk := processor.NewHistogramProcessor(config).WithInterceptor(s).WithLogger(s.log)

	// process incoming metrics
	if err := bulk.ReadFromStream(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get remote endpoint with passed query args
	u, _ := url.Parse(endpoint)
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

// func copyHeaders(dst, src http.Header) {
// 	for k, vv := range src {
// 		for _, v := range vv {
// 			dst.Add(k, v)
// 		}
// 	}
// }

// func httpForwarder(w http.ResponseWriter, r *http.Request) {
// 	res, err := http.DefaultTransport.RoundTrip(r)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}
// 	defer res.Body.Close()
// 	copyHeaders(w.Header(), res.Header)
// 	w.WriteHeader(res.StatusCode)
// 	if _, err := io.Copy(w, res.Body); err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}
// }

// func (s *Service) streamResponse() {
// 	httpClient := &http.Client{}
// 	req, _ := http.NewRequest(httpMethod, fmt.Sprintf("%s/?region=%s", cfg.Endpoint, record.AWSRegion), obj.Body)
// 	httpUserAgent := fmt.Sprintf("kourier-lambda/%s", cfg.Version)
// 	referer := fmt.Sprintf("s3://%s/%s", record.S3.Bucket.Name, record.S3.Object.Key)
// 	req.Header.Set("Content-Encoding", httpContentEncoding)
// 	req.Header.Set("User-Agent", httpUserAgent)
// 	req.Header.Set("Referer", referer)
// 	res, err := httpClient.Do(req)
// 	if err != nil {
// 		return errors.Wrap(err, "http")

// 	}
// 	defer res.Body.Close()
// 	if res.StatusCode != http.StatusOK {
// 		return errors.Wrap(fmt.Errorf("invalid response status - %d", res.StatusCode), "http")
// 	}
// 	return nil

// }

// convert http event to obfuscated metric
func (s *Service) OnMetric(m *metric.Metric) error {
	// ensure "url" and "method" labels exist in the source metric
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

	// add default labels
	m.Labels["method"] = method
	m.Labels["host"] = u.Host
	m.Labels["scheme"] = u.Scheme
	m.Labels["operation_id"] = ""
	m.Labels["name"] = ""
	m.Labels["tag"] = ""

	// search openapi schemas for specified url
	req := http.Request{
		Method: method,
		URL:    u,
	}
	route, err := s.api.Resolve(req)

	// add openapi-specific labels
	switch err {
	case nil:
		m.Labels["operation_id"] = route.OperationID
		m.Labels["name"] = route.SpecID
		m.Labels["tag"] = route.Tag

	case openapi.ErrRouteNotFound:
		s.log.Warn("No route found, keeping defaults")
		err = nil
	}
	return err
}
