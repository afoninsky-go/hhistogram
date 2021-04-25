package openapi

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// https://demotenant.dev.mambucloud.com/apidocs/

func TestCommon(t *testing.T) {
	p := NewURLParser()
	p.LoadFromFile("test-tenant", "./swagger.json", []string{"api.mambu.com"})
	// p.LoadFromFile("default", "./swagger.json", []string{})

	reqURL, _ := url.Parse("http://api.mamsbu.com/clients/clientid")
	res, err := p.FindRoute(http.MethodGet, *reqURL)
	fmt.Println(res)
	fmt.Println(err)

	// swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("./swagger.json")
	// if err != nil {
	// 	panic(err)
	// }
	// router, _ := gorillamux.NewRouter(swagger)

	// reqURL := &url.URL{
	// 	Scheme: "http",
	// 	Path:   "/clients/hello/creditarrangements",
	// }
	// req := http.Request{
	// 	Method: "GET",
	// 	URL:    reqURL,
	// }

	// // ErrPathNotFound
	// // ErrMethodNotAllowed
	// route, _, err := router.FindRoute(&req)

	// if err == nil {
	// 	fmt.Println("path", route.Path)
	// 	fmt.Println("method", route.Method)
	// 	fmt.Println(route)
	// } else {
	// 	fmt.Println(err)
	// 	fmt.Println(err == routers.ErrMethodNotAllowed)
	// 	fmt.Println(err == routers.ErrPathNotFound)
	// }
}
