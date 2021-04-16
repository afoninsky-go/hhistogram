package processor

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/logger"
)

type handler struct{}

func (s *handler) OnMetric(m *metric.Metric) error {
	m.SetName("handled")
	fmt.Println(m)
	return nil
}

func TestCommon(t *testing.T) {
	log := logger.NewSTDLogger()

	metrics, err := os.Open("./debug.txt")
	log.FatalIfError(err)

	cfg := NewHistogramConfig().WithName("test")
	p := NewHistogramProcessor(*cfg).WithInterceptor(&handler{})

	log.FatalIfError(p.ReadFromStream(metrics))

	var buf bytes.Buffer
	p.Process(&buf)

	fmt.Println(buf)
}
