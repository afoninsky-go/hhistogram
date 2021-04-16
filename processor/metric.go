package processor

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

	Name      string `json:"-"`
	Validated bool   `json:"-"`
}

// create metric from JSON representation:
// {"metric":{"__name__":"requests_total","instance":"localhost","port":"9090"},"values":[123],"timestamps":[1598089314604]}
func (s *Metric) UnmarshalJSON(buf []byte) error {
	if err := json.Unmarshal(buf, s); err != nil {
		return err
	}
	return s.Validate()
}

// get readable metric name:
// requests_total{instance="localhost",port="9090"}
func (s *Metric) String() string {
	if len(s.Labels) == 0 {
		return s.Name
	}
	pairs := []string{}
	for k, v := range s.Labels {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, v))
	}
	return fmt.Sprintf("%s{%s}", s.Name, strings.Join(pairs, ","))
}

// generate a set of metrics with one timestamp/value from the generic one
func (s *Metric) Slice() []Metric {
	items := []Metric{}
	for i, _ := range s.Values {
		items = append(items, Metric{
			Name:       s.Name,
			Validated:  s.Validated,
			Labels:     s.Labels,
			Values:     []float64{s.Values[i]},
			Timestamps: []int64{s.Timestamps[i]},
		})
	}
	return items
}

// ensure metric is valid
func (s *Metric) Validate() error {
	if _, ok := s.Labels[metricNameLabel]; !ok {
		return errors.New("name label does not exit")
	}
	s.Name = s.Labels[metricNameLabel]
	delete(s.Labels, metricNameLabel)

	if len(s.Values) != len(s.Timestamps) {
		return errors.New("amoint of values is not equal to timestamps")
	}
	if len(s.Values) == 0 {
		return errors.New("metric is empty")
	}
	s.Validated = true
	return nil
}
