package openapi

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"os"
// 	"path"
// 	"path/filepath"
// 	"strings"

// 	"github.com/afoninsky-go/logger"
// 	"github.com/getkin/kin-openapi/openapi3"
// 	"github.com/getkin/kin-openapi/routers"
// 	"github.com/getkin/kin-openapi/routers/gorillamux"
// 	"github.com/gorilla/mux"
// 	lru "github.com/hashicorp/golang-lru"
// )

// const defaultHost = "localhost"
// const lruCacheSize = 1e6

// var ErrRouteNotFound = errors.New("route not found")

// type routerSpec struct {
// 	routers.Router
// 	// identifier of opeanapi spec which implements this router
// 	SpecID string
// }

// type cacheSpec struct {
// 	Route OpenAPIRoute
// 	Err   error
// }

// type OpenAPI struct {
// 	routers map[string][]routerSpec
// 	log     *logger.Logger
// 	cache   *lru.Cache

// 	muxRouter *mux.Router
// }

// // router found based on passed url
// type OpenAPIRoute struct {
// 	// identifier of openapi spec where selected route was found
// 	SpecID string
// 	// openapi http path
// 	Path string
// 	// openapi route operation id (optional)
// 	OperationID string
// 	// openapi route tag (optional)
// 	Tag string
// }

// func NewURLParser() *OpenAPI {
// 	s := &OpenAPI{}
// 	s.routers = map[string][]routerSpec{}
// 	s.log = logger.NewSTDLogger()
// 	cache, _ := lru.New(lruCacheSize)
// 	s.cache = cache

// 	s.muxRouter = mux.NewRouter()
// 	return s
// }

// // load swagger specs from folder
// func (s *OpenAPI) LoadFolder(folder string, hosts []string) error {
// 	files, err := ioutil.ReadDir(folder)
// 	if err != nil {
// 		return err
// 	}

// 	for _, f := range files {
// 		fileName := f.Name()
// 		filePath := path.Join(folder, fileName)
// 		fileID := strings.TrimSuffix(fileName, filepath.Ext(fileName))
// 		if err := s.AddSpec(fileID, filePath, hosts); err != nil {
// 			return err
// 		}
// 		s.log.Info("Spec loaded: ", fileName)
// 	}

// 	return nil
// }

// // load swagger specification from file
// func (s *OpenAPI) AddSpec(specID, path string, hosts []string) error {
// 	spec, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(path)
// 	if err != nil {
// 		return err
// 	}
// 	if err := spec.Validate(context.Background()); err != nil {
// 		return err
// 	}
// 	fmt.Println(s.getBasePath(spec))
// 	os.Exit(0)

// 	// create routers
// 	apiRouter, _ := gorillamux.NewRouter(spec)
// 	router := routerSpec{SpecID: specID, Router: apiRouter}

// 	// setup default host if not hosts specified
// 	if len(hosts) == 0 {
// 		hosts = []string{defaultHost}
// 	}

// 	for _, host := range hosts {
// 		s.routers[host] = append(s.routers[host], router)
// 	}

// 	return nil
// }

// func (s *OpenAPI) getBasePath(spec *openapi3.Swagger) string {

// 	buf, err := spec.MarshalJSON()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(string(buf))
// 	return "qwe"
// }

// func (s *OpenAPI) WithLogger(log *logger.Logger) *OpenAPI {
// 	s.log = log
// 	return s
// }

// // search request in loaded openapi specification
// func (s *OpenAPI) Resolve(req http.Request) (OpenAPIRoute, error) {
// 	cacheID := fmt.Sprintf("%s|%s", req.Method, req.URL)

// 	// check if url already cached
// 	if s.cache != nil {
// 		cache, ok := s.cache.Get(cacheID)
// 		if ok {
// 			res := cache.(cacheSpec)
// 			return res.Route, res.Err
// 		}
// 	}

// 	// resolve url against loaded openapi specifications
// 	path, err := s.findRoute(req)
// 	if err != nil {
// 		// not found in specific host -> try to find in default one
// 		if err == ErrRouteNotFound && req.URL.Host != defaultHost {
// 			req.URL.Host = defaultHost
// 			path, err = s.findRoute(req)
// 		}
// 	}

// 	// cache result in order to speedup next requests
// 	if s.cache != nil {
// 		s.cache.Add(cacheID, cacheSpec{
// 			Route: path,
// 			Err:   err,
// 		})
// 	}

// 	return path, err
// }

// func (s *OpenAPI) findRoute(req http.Request) (OpenAPIRoute, error) {
// 	result := OpenAPIRoute{}
// 	routersSlice, exists := s.routers[req.URL.Host]
// 	if !exists {
// 		return result, ErrRouteNotFound
// 	}

// 	var resultErr error
// 	for _, router := range routersSlice {
// 		route, _, err := router.FindRoute(&req)
// 		resultErr = err
// 		switch err {
// 		case nil:
// 			// return matched route with operation id and tag if exists
// 			result.SpecID = router.SpecID
// 			result.Path = route.Path
// 			if route.Operation != nil {
// 				result.OperationID = route.Operation.OperationID
// 				if len(route.Operation.Tags) > 0 {
// 					result.Tag = strings.ToLower(route.Operation.Tags[0])
// 				}
// 			}
// 			return result, resultErr
// 		case routers.ErrMethodNotAllowed:
// 			resultErr = ErrRouteNotFound
// 			break
// 		case routers.ErrPathNotFound:
// 			resultErr = ErrRouteNotFound
// 			break
// 		default:
// 			break
// 		}
// 	}
// 	return result, resultErr
// }
