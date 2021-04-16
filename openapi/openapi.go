package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const defaultHost = "localhost"
const lruCacheSize = 100000

var ErrRouteNotFound = errors.New("route not found")

type OpenAPI struct {
	routers map[string]routers.Router
}

type OpenAPIRoute struct {
	OperationID string
	Tag         string
	Path        string
}

func NewURLParser() *OpenAPI {
	s := &OpenAPI{}
	s.routers = make(map[string]routers.Router, 0)

	return s
}

// load swagger specification from file
func (s *OpenAPI) LoadFromFile(path string, hosts []string) error {
	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("./swagger.json")
	if err != nil {
		return err
	}
	// create routers
	router, _ := gorillamux.NewRouter(spec)
	if len(hosts) > 0 {
		for _, host := range hosts {
			s.routers[host] = router
		}
	} else {
		s.routers[defaultHost] = router
	}
	return nil
}

func (s *OpenAPI) FindRoute(method string, reqURL url.URL) (*OpenAPIRoute, error) {
	// TODO: cache find route requests

	var path *OpenAPIRoute
	var err error

	// try to find route with specified host
	path, err = s.findRoute(method, &reqURL)
	if err != nil {
		// not found in specific host -> try to find in default one
		if err == ErrRouteNotFound && reqURL.Host != defaultHost {
			reqURL.Host = defaultHost
			path, err = s.findRoute(method, &reqURL)
		}
	}
	return path, err
}

func (s *OpenAPI) findRoute(method string, reqURL *url.URL) (*OpenAPIRoute, error) {
	req := http.Request{
		Method: method,
		URL:    reqURL,
	}
	router, exists := s.routers[reqURL.Host]
	if !exists {
		return nil, ErrRouteNotFound
	}
	route, _, err := router.FindRoute(&req)
	if err != nil {
		switch err {
		case routers.ErrMethodNotAllowed:
			return nil, ErrRouteNotFound
		case routers.ErrPathNotFound:
			return nil, ErrRouteNotFound
		default:
			return nil, err
		}
	}

	// return matched route with operation id and tag if exists
	res := &OpenAPIRoute{}
	res.Path = route.Path
	if route.Operation != nil {
		res.OperationID = route.Operation.OperationID
		if len(route.Operation.Tags) > 0 {
			res.Tag = route.Operation.Tags[0]
		}
	}

	return res, nil
}

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
