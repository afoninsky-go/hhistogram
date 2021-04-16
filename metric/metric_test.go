package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ethalon = `{"metric":{"__name__":"requests_total","instance":"localhost","port":"9090"},"values":[123, 456],"timestamps":[1598089314604, 1598089314604]}`

func TestCommon(t *testing.T) {
	// test create metric from json
	var m Metric
	if err := m.UnmarshalJSON([]byte(ethalon)); err != nil {
		panic(err)
	}
	assert.Equal(t, `requests_total{instance="localhost",port="9090"}`, m.String())
	assert.Equal(t, 2, len(m.Values))

	// test update metric name
	m.SetName("test")
	assert.Equal(t, `test{instance="localhost",port="9090"}`, m.String())

	// test slice method
	slices := m.Slice()
	assert.Equal(t, 2, len(slices))
}
