package openapi

import (
	"context"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"
)

const lruCacheSize = 1e6

type swagger struct {
	routes     []*mux.Route
	routeSpecs []RouteSpec
	cache      *lru.Cache
}

type RouteSpec struct {
	// identifier of openapi spec where selected route was found
	ID string
	// openapi http path
	Path string
	// openapi route operation id (optional)
	OperationID string
	// openapi route tag (optional)
	Tag string
}

func NewSwaggerRouter() *swagger {
	s := swagger{}
	cache, _ := lru.New(lruCacheSize)
	s.cache = cache
	return &s
}

// load swagger specs from folder
func (s *swagger) LoadSpecFolder(folder string) error {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return err
	}

	for _, f := range files {
		fileName := f.Name()
		filePath := path.Join(folder, fileName)
		fileID := tokenize(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		if err := s.AddSpecFromFile(fileID, filePath); err != nil {
			return err
		}
	}

	return nil
}

// loads routes from swagger specification
func (s *swagger) AddSpecFromFile(id, path string) error {
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
			urlPath := spec.BasePath + routeName
			muxRouter := mux.NewRouter().HandleFunc(urlPath, dummyHandler).
				Schemes(spec.Schemes...).
				Methods(method)
			s.routes = append(s.routes, muxRouter)
			routeSpec := RouteSpec{
				ID:   id,
				Path: urlPath,
			}
			if len(operation.Tags) > 0 {
				routeSpec.Tag = tokenize(operation.Tags[0])
			}
			if operation.OperationID != "" {
				routeSpec.OperationID = tokenize(operation.OperationID)
			}
			s.routeSpecs = append(s.routeSpecs, routeSpec)
		}
	}
	return nil
}

// tests request against loaded routes
func (s *swagger) TestRoute(req *http.Request) *RouteSpec {
	cacheID := s.requestID(req)
	cache, ok := s.cache.Get(cacheID)
	if ok {
		return s.getRouteSpec(cache.(int))
	}
	index := s.getRouteIndex(req)
	s.cache.Add(cacheID, index)
	return s.getRouteSpec(index)
}

// generates unique request identificator based on configured matchers
// used for caching purposes mostly
func (s *swagger) requestID(req *http.Request) string {
	if req == nil {
		return ""
	}
	parts := []string{req.Method, req.URL.Path}
	return strings.Join(parts, ",")
}

func (s *swagger) getRouteIndex(req *http.Request) int {
	for i, route := range s.routes {
		var match mux.RouteMatch
		if route.Match(req, &match) && match.MatchErr == nil {
			return i
		}
	}
	return -1
}

func (s *swagger) getRouteSpec(i int) *RouteSpec {
	if i < 0 {
		return nil
	}
	if i > len(s.routeSpecs)-1 {
		return nil
	}
	return &s.routeSpecs[i]
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi!"))
}

func tokenize(token string) string {
	result := ""
	for _, i := range token {
		if unicode.IsLetter(i) {
			result = result + string(i)
		}
	}
	return strings.ToLower(result)
}
