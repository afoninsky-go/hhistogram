package processor

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/logger"
	"github.com/stretchr/testify/assert"
)

const inputMetrics = `
{"metric":{"__name__":"http_request_latency","method":"POST"},"values":[1,2],"timestamps":[1598089314604,1598089314604]}
{"metric":{"__name__":"http_request_latency","method":"GET"},"values":[5],"timestamps":[1598089314604]}
`
const outputMetrics = `
test_processed_bucket{method="GET",vmrange="4.642e+00...5.275e+00"} 1
test_processed_sum{method="GET"} 5
test_processed_count{method="GET"} 1
test_processed_bucket{method="POST",vmrange="8.799e-01...1.000e+00"} 1
test_processed_bucket{method="POST",vmrange="1.896e+00...2.154e+00"} 1
test_processed_sum{method="POST"} 3
test_processed_count{method="POST"} 2
`

func absPath(path string) string {
	folder, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return folder
}

type handler struct{}

func (s *handler) OnMetric(m *metric.Metric) error {
	m.SetName(m.GetName() + "_processed")
	return nil
}

func Test500kEvents(t *testing.T) {
	log := logger.NewSTDLogger()
	cfg := NewConfig().WithName("test")
	p := NewHistogramProcessor(cfg).WithInterceptor(&handler{})

	r := strings.NewReader(inputMetrics)
	log.FatalIfError(p.ReadFromStream(r))

	var buf bytes.Buffer
	p.Process(&buf)

	assert.Equal(t, strings.TrimSpace(outputMetrics), strings.TrimSpace(buf.String()))
}
