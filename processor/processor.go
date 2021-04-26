package processor

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/logger"
)

type MetricInterceptor interface {
	OnMetric(*metric.Metric) error
}

type Processor struct {
	sync.Mutex
	cfg            Config
	buckets        map[time.Time][]metric.Metric
	interceptor    MetricInterceptor
	log            *logger.Logger
	processedCount uint32
	inputCount     uint32
}

func NewHistogramProcessor(cfg Config) *Processor {
	p := &Processor{}
	p.cfg = cfg
	p.buckets = make(map[time.Time][]metric.Metric, 0)
	p.log = logger.NewSTDLogger()
	return p
}

// reads metrics in json format from the stream
func (s *Processor) ReadFromStream(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s.inputCount++
		var m metric.Metric
		buf := scanner.Bytes()
		if len(buf) == 0 {
			continue
		}
		if err := m.UnmarshalJSON(buf); err != nil {
			fmt.Println(err, string(buf))
			continue
		}
		for _, m1 := range m.Slice() {
			// create metric with one dimension and redefine its name
			m1.SetName(s.cfg.Name)
			if s.interceptor != nil {
				if err := s.interceptor.OnMetric(&m1); err != nil {
					return err
				}
			}
			if m1.GetName() != "" {
				s.pushToBucket(m1)
			}
		}
	}
	return scanner.Err()
}

// specify metric handler for every incoming metric
func (s *Processor) WithInterceptor(handler MetricInterceptor) *Processor {
	s.interceptor = handler
	return s
}

func (s *Processor) WithLogger(log *logger.Logger) *Processor {
	s.log = log
	return s
}

// creates metrics related to histograms based on buckets
// we expect that buckets contain one-dimension metrics created by metric.Slice()
func (s *Processor) Process(w io.Writer) {
	s.Lock()
	defer s.Unlock()
	for _, events := range s.buckets {
		// create a slice of histogram with metrics
		storage := metrics.NewSet()
		for _, event := range events {
			name := event.String()
			for _, value := range event.Values {
				s.processedCount++
				storage.GetOrCreateHistogram(name).Update(value)
			}
		}
		// read resulting metrics
		storage.WritePrometheus(w)
	}
	s.log.Infof("Processed %d of %d events", s.processedCount, s.inputCount)
}

// places metrics based on their timestamps to a specific bucket
func (s *Processor) pushToBucket(m metric.Metric) {
	s.Lock()
	defer s.Unlock()
	timestamp := time.Unix(0, m.Timestamps[0]*int64(time.Millisecond))
	upperBorder := timestamp.Truncate(s.cfg.SliceDuration).Add(s.cfg.SliceDuration)
	if _, ok := s.buckets[upperBorder]; !ok {
		s.buckets[upperBorder] = []metric.Metric{}
	}
	s.buckets[upperBorder] = append(s.buckets[upperBorder], m)
}
