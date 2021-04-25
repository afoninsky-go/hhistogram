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
	p.AddSpec("test-tenant", "./swagger.json", []string{"api.mambu.com"})
	p.AddSpec("default-tenant", "./swagger.json", []string{"localhost"})

	reqURL, _ := url.Parse("http://api.dmambu.com/clients/clientid")
	req := http.Request{
		Method: http.MethodGet,
		URL:    reqURL,
	}

	res1, err1 := p.Resolve(req)
	res2, err2 := p.Resolve(req)
	fmt.Println(res1)
	fmt.Println(res2)
	fmt.Println(err1)
	fmt.Println(err2)

}
