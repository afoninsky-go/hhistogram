package processor

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type Processor struct {
	sync.Mutex
	cfg     Config
	buckets map[time.Time][]Metric
}

func NewHistogramProcessor(cfg Config) *Processor {
	p := &Processor{}
	p.cfg = cfg
	p.buckets = make(map[time.Time][]Metric, 0)
	return p
}

// reads metrics in json format from the stream
func (s *Processor) AppendFromStream(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var metric Metric
		if err := metric.UnmarshalJSON(scanner.Bytes()); err != nil {
			return err
		}
		for _, m := range metric.Slice() {
			// create metric with one dimension and redefine its name
			m.Name = s.cfg.Name
			s.PushToBucket(m)
		}
	}
	return scanner.Err()
}

// places metrics based on their timestamps to a specific bucket
func (s *Processor) PushToBucket(m Metric) {
	s.Lock()
	defer s.Unlock()
	timestamp := time.Unix(0, m.Timestamps[0]*int64(time.Millisecond))
	upperBorder := timestamp.Truncate(s.cfg.SliceDuration).Add(s.cfg.SliceDuration)
	if _, ok := s.buckets[upperBorder]; !ok {
		s.buckets[upperBorder] = []Metric{}
	}
	s.buckets[upperBorder] = append(s.buckets[upperBorder], m)
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
				storage.GetOrCreateHistogram(name).Update(value)
			}
		}
		// read resulting metrics
		storage.WritePrometheus(w)
	}
}
