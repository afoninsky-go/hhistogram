package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/afoninsky-go/logger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	lru "github.com/hashicorp/golang-lru"
)

const defaultHost = "localhost"
const lruCacheSize = 1e6

var ErrRouteNotFound = errors.New("route not found")

type routerSpec struct {
	routers.Router
	// identifier of opeanapi spec which implements this router
	SpecID string
}

type cacheSpec struct {
	Route OpenAPIRoute
	Err   error
}

type OpenAPI struct {
	routers map[string]routerSpec
	log     *logger.Logger
	cache   *lru.Cache
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
	cache, _ := lru.New(lruCacheSize)
	s.cache = cache
	return s
}

// load swagger specification from file
func (s *OpenAPI) AddSpec(specID, path string, hosts []string) error {
	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(path)
	if err != nil {
		return err
	}
	// create routers
	apiRouter, _ := gorillamux.NewRouter(spec)
	router := routerSpec{SpecID: specID, Router: apiRouter}
	if len(hosts) == 0 {
		return errors.New("no hosts specified")
	}
	for _, host := range hosts {
		if _, exists := s.routers[host]; exists {
			return fmt.Errorf("host already loaded: %s", host)
		}
		s.routers[host] = router
	}

	return nil
}

func (s *OpenAPI) WithLogger(log *logger.Logger) *OpenAPI {
	s.log = log
	return s
}

// search request in loaded openapi specification
func (s *OpenAPI) Resolve(req http.Request) (OpenAPIRoute, error) {
	cacheID := fmt.Sprintf("%s|%s", req.Method, req.URL)

	// check if url already cached
	if s.cache != nil {
		cache, ok := s.cache.Get(cacheID)
		if ok {
			res := cache.(cacheSpec)
			return res.Route, res.Err
		}
	}

	// resolve url against loaded openapi specifications
	path, err := s.findRoute(req)
	if err != nil {
		// not found in specific host -> try to find in default one
		if err == ErrRouteNotFound && req.URL.Host != defaultHost {
			req.URL.Host = defaultHost
			path, err = s.findRoute(req)
		}
	}

	// cache result in order to speedup next requests
	if s.cache != nil {
		s.cache.Add(cacheID, cacheSpec{
			Route: path,
			Err:   err,
		})
	}

	return path, err
}

func (s *OpenAPI) findRoute(req http.Request) (OpenAPIRoute, error) {
	result := OpenAPIRoute{}
	router, exists := s.routers[req.URL.Host]
	if !exists {
		return result, ErrRouteNotFound
	}
	route, _, err := router.FindRoute(&req)
	if err != nil {
		switch err {
		case routers.ErrMethodNotAllowed:
			return result, ErrRouteNotFound
		case routers.ErrPathNotFound:
			return result, ErrRouteNotFound
		default:
			return result, err
		}
	}

	// return matched route with operation id and tag if exists
	result.SpecID = router.SpecID
	result.Path = route.Path
	if route.Operation != nil {
		result.OperationID = route.Operation.OperationID
		if len(route.Operation.Tags) > 0 {
			result.Tag = strings.ToLower(route.Operation.Tags[0])
		}
	}

	return result, nil
}
