package openapi

import (
	"fmt"
	"net/http"
	"testing"
)

// https://demotenant.dev.mambucloud.com/apidocs/

func TestMambuSpecs(t *testing.T) {
	p := NewSwaggerRouter()
	if err := p.AddSpecFromFile("../test/json/users_v2_swagger.json"); err != nil {
		panic(err)
	}
	if err := p.AddSpecFromFile("../test/json/configuration__branches.yaml_v2_swagger.json"); err != nil {
		panic(err)
	}
	fmt.Println(">>>>")
	req1, _ := http.NewRequest(http.MethodGet, "http://localhost/api/users", nil)
	req2, _ := http.NewRequest(http.MethodGet, "http://localhost/api/users", nil)
	res1 := p.TestRoute(req1)
	res2 := p.TestRoute(req2)
	fmt.Println(res1)
	fmt.Println(res2)
	fmt.Println(res1 == res2)
}

// func TestMambuSpecs(t *testing.T) {
// 	p := NewURLParser()
// 	folder, err := filepath.Abs("../test/json")
// 	assert.NoError(t, err)
// 	assert.NoError(t, p.LoadFolder(folder, []string{}))

// 	u, _ := url.Parse("https://obkesp.mambu.com:443/api/deposits/LSA143884679VE?detailsLevel=FULL")
// 	req := http.Request{
// 		Method: http.MethodGet,
// 		URL:    u,
// 	}
// 	res, err := p.Resolve(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(res)
// }

// func TestCommon(t *testing.T) {
// 	p := NewURLParser()

// 	// load specs
// 	if err := p.AddSpec("users-spec", absPath("../test/json/users_v2_swagger.json"), []string{}); err != nil {
// 		t.Error(err)
// 	}
// 	if err := p.AddSpec("clients-spec", absPath("../test/json/clients_v2_swagger.json"), []string{}); err != nil {
// 		t.Error(err)
// 	}
// 	if err := p.AddSpec("loans-spec", absPath("../test/json/loans_transactions_v2_swagger.json"), []string{}); err != nil {
// 		t.Error(err)
// 	}

// 	// returs 404 for non-existing route in specs
// 	assert.EqualError(t,
// 		assertRoute(t, p, http.MethodGet, "http://api.mambu.com", "", "", "", ""),
// 		ErrRouteNotFound.Error(),
// 	)

// 	// resolves client route
// 	assert.NoError(t,
// 		assertRoute(t, p, http.MethodGet, "http://api.mambu.com/clients/id", "clients-spec", "/clients/{clientId}", "getById", "clients"),
// 	)
// 	// returs 404 for the same route but another method
// 	assert.EqualError(t,
// 		assertRoute(t, p, http.MethodPost, "http://api.mambu.com/clients/id", "", "", "", ""),
// 		ErrRouteNotFound.Error(),
// 	)

// 	// resolves loan route
// 	assert.NoError(t,
// 		assertRoute(t, p, http.MethodPost, "http://localhost/loans/id/lock-transactions", "loans-spec", "/loans/{loanAccountId}/lock-transactions", "applyLock", "loantransactions"),
// 	)

// 	// check caching function
// 	assert.NoError(t,
// 		assertRoute(t, p, http.MethodPost, "http://localhost/loans/id/lock-transactions", "loans-spec", "/loans/{loanAccountId}/lock-transactions", "applyLock", "loantransactions"),
// 	)

// }
