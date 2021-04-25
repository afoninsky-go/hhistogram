package openapi

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/afoninsky-go/hhistogram/metric"
	"github.com/afoninsky-go/logger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const defaultHost = "localhost"
const lruCacheSize = 100000

var ErrRouteNotFound = errors.New("route not found")

type routerSpec struct {
	routers.Router
	// identifier of opeanapi spec which implements this router
	SpecID string
}
type OpenAPI struct {
	routers map[string]routerSpec
	log     *logger.Logger
}

// router found based on passed url
type OpenAPIRoute struct {
	// identifier of openapi spec where selected route was found
	SpecID string
	// openapi http path
	Path string
	// openapi route operation id (optional)
	OperationID string
	// openapi route tag (optional)
	Tag string
}

func NewURLParser() *OpenAPI {
	s := &OpenAPI{}
	s.routers = make(map[string]routerSpec, 0)
	s.log = logger.NewSTDLogger()
	return s
}

// load swagger specification from file
func (s *OpenAPI) LoadFromFile(specID, path string, hosts []string) error {
	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("./swagger.json")
	if err != nil {
		return err
	}
	// create routers
	apiRouter, _ := gorillamux.NewRouter(spec)
	router := routerSpec{SpecID: specID, Router: apiRouter}
	if len(hosts) > 0 {
		for _, host := range hosts {
			s.routers[host] = router
		}
	} else {
		s.routers[defaultHost] = router
	}
	return nil
}

func (s *OpenAPI) WithLogger(log *logger.Logger) *OpenAPI {
	s.log = log
	return s
}

func (s *OpenAPI) FindRoute(method string, reqURL url.URL) (*OpenAPIRoute, error) {
	// TODO: cache find route requests

	var path *OpenAPIRoute
	var err error

	// try to find route with specified host
	path, err = s.findRoute(method, reqURL)
	if err != nil {
		// not found in specific host -> try to find in default one
		if err == ErrRouteNotFound && reqURL.Host != defaultHost {
			reqURL.Host = defaultHost
			path, err = s.findRoute(method, reqURL)
		}
	}
	return path, err
}

func (s *OpenAPI) findRoute(method string, reqURL url.URL) (*OpenAPIRoute, error) {
	req := http.Request{
		Method: method,
		URL:    &reqURL,
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
	res := &OpenAPIRoute{
		SpecID: router.SpecID,
		Path:   route.Path,
	}
	if route.Operation != nil {
		res.OperationID = route.Operation.OperationID
		if len(route.Operation.Tags) > 0 {
			res.Tag = strings.ToLower(route.Operation.Tags[0])
		}
	}

	return res, nil
}

// checks if metric has url-specific labels and decreases its cardinality
func (s *OpenAPI) OnMetric(m *metric.Metric) error {
	// ensure passed metric has "url" label
	rawurl, ok := m.Labels["url"]
	if !ok {
		s.log.Warn("metric doesn't have url label, ignoring ...")
		return nil
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		s.log.Warnf("unable to parse %s, ignoring...", rawurl)
		return nil
	}

	// remove high-cardinality url label
	delete(m.Labels, "url")

	// get method or set default
	method := http.MethodGet
	_method, ok := m.Labels["method"]
	if ok {
		method = _method
	}

	// add default labels
	m.Labels["method"] = method
	m.Labels["host"] = u.Host
	m.Labels["scheme"] = u.Scheme
	m.Labels["operation_id"] = ""
	m.Labels["name"] = ""
	m.Labels["tag"] = ""

	// search loaded openapi schemas for specified url
	route, err := s.FindRoute(method, *u)

	switch err {
	case nil:
		m.Labels["operation_id"] = route.OperationID
		m.Labels["name"] = route.SpecID
		m.Labels["tag"] = route.Tag

	case ErrRouteNotFound:
		s.log.Warn("No route found, keeping defaults")
		err = nil
	}
	return err
}
