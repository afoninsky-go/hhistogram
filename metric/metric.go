package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const metricNameLabel = "__name__"

type Metric struct {
	Labels map[string]string `json:"metric"`
	Values []float64         `json:"values"`
	// milliseconds from unix epoch
	Timestamps []int64 `timestamps:"timestamps"`

	name      string
	validated bool
}

type rawMetric struct {
	Labels map[string]string `json:"metric"`
	Values []float64         `json:"values"`
	// milliseconds from unix epoch
	Timestamps []int64 `timestamps:"timestamps"`
}

// create metric from JSON representation:
// {"metric":{"__name__":"requests_total","instance":"localhost","port":"9090"},"values":[123],"timestamps":[1598089314604]}
func (s *Metric) UnmarshalJSON(buf []byte) error {
	var m rawMetric
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	s.metricFromRaw(m)
	return s.validate()
}

func (s *Metric) SetName(name string) *Metric {
	s.name = name
	return s
}

// get readable metric name:
// requests_total{instance="localhost",port="9090"}
func (s *Metric) String() string {
	if len(s.Labels) == 0 {
		return s.name
	}
	pairs := []string{}
	for k, v := range s.Labels {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return fmt.Sprintf("%s{%s}", s.name, strings.Join(pairs, ","))
}

// generate a set of metrics with one timestamp/value from the generic one
func (s *Metric) Slice() []Metric {
	items := []Metric{}
	for i, _ := range s.Values {
		items = append(items, Metric{
			name:       s.name,
			Labels:     s.Labels,
			Values:     []float64{s.Values[i]},
			Timestamps: []int64{s.Timestamps[i]},
		})
	}
	return items
}

// ensure metric is valid
func (s *Metric) validate() error {
	if len(s.Values) != len(s.Timestamps) {
		return errors.New("amount of values is not equal to timestamps")
	}
	if len(s.Values) == 0 {
		return errors.New("metric is empty")
	}
	return nil
}

func (s *Metric) metricFromRaw(m rawMetric) {
	s.Labels = m.Labels
	s.Values = m.Values
	s.Timestamps = m.Timestamps
	s.name = m.Labels[metricNameLabel]
	delete(m.Labels, metricNameLabel)
}
