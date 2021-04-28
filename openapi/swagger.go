package openapi

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
)

const lruCacheSize = 1e6

type swagger struct {
	routes     []*mux.Route
	operations []*openapi2.Operation
	cache      *lru.Cache
}

func NewSwaggerRouter() *swagger {
	s := swagger{}
	cache, _ := lru.New(lruCacheSize)
	s.cache = cache
	return &s
}

// generates unique request identificator based on configured matchers
func (s *swagger) requestID(req *http.Request) string {
	if req == nil {
		return ""
	}
	parts := []string{req.Method, req.URL.Path}
	return strings.Join(parts, ",")
}

// loads routes from swagger specification
func (s *swagger) addSpecFromFile(path string) error {
	spec := openapi2.Swagger{}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := spec.UnmarshalJSON(buf); err != nil {
		return err
	}
	if err := spec.Info.Validate(context.Background()); err != nil {
		return err
	}
	for routeName, routePath := range spec.Paths {
		for method, operation := range routePath.Operations() {
			muxRouter := mux.NewRouter().HandleFunc(spec.BasePath+routeName, dummyHandler).
				Schemes(spec.Schemes...).
				Methods(method)
			s.routes = append(s.routes, muxRouter)
			s.operations = append(s.operations, operation)
		}
	}
	return nil
}

// tests request against loaded routes
func (s *swagger) testRoute(req *http.Request) *openapi2.Operation {
	// check if route is in cache
	cacheID := s.requestID(req)
	cache, ok := s.cache.Get(cacheID)
	if ok {
		return s.getOperationByIndex(cache.(int))
	}

	// search route in routing table
	index := -1
	for i, route := range s.routes {
		var match mux.RouteMatch
		if route.Match(req, &match) {
			switch match.MatchErr {
			case nil:
				index = i
				// case mux.ErrMethodMismatch:
				// case mux.ErrNotFound:
			}
		}
	}
	s.cache.Add(cacheID, index)
	return s.getOperationByIndex(index)
}

func (s *swagger) getOperationIndex(req *http.Request) int {
	for i, route := range s.routes {
		var match mux.RouteMatch
		if route.Match(req, &match) && match.MatchErr == nil {
			return i
		}
	}
	return -1
}

func (s *swagger) getOperationByIndex(i int) *openapi2.Operation {
	if i < 0 {
		return nil
	}
	if i > len(s.operations)-1 {
		return nil
	}
	return s.operations[i]
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi!"))
}
