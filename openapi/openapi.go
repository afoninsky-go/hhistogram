package openapi

import (
	"fmt"

	"github.com/afoninsky-go/hhistogram/metric"
)

type OpenAPI struct {
}

func NewURLParser() *OpenAPI {

	return &OpenAPI{}
}

// https://swagger.io/specification/
// TODO: route using:
// - servers (url)
// - paths

// checks if metric has url-specific labels and decreases its cardinality
func (s *OpenAPI) OnMetric(m *metric.Metric) error {
	// TODO: https://golang.org/pkg/net/url/#example_URL
	// - create url from labels
	// - pass to openapi router
	// - update fields based on results
	m.SetName("handled")
	fmt.Println(m)
	return nil
}
