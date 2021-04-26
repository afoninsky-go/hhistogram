package openapi

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// https://demotenant.dev.mambucloud.com/apidocs/

func assertRoute(t *testing.T, p *OpenAPI, method, reqURL, specID, path, operationID, tag string) error {
	u, _ := url.Parse(reqURL)
	req := http.Request{
		Method: method,
		URL:    u,
	}
	res, err := p.Resolve(req)
	if err != nil {
		return err
	}
	assert.Equal(t, specID, res.SpecID)
	assert.Equal(t, path, res.Path)
	assert.Equal(t, operationID, res.OperationID)
	assert.Equal(t, tag, res.Tag)

	return nil
}

func absPath(path string) string {
	folder, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return folder
}

func TestFolder(t *testing.T) {
	p := NewURLParser()
	folder, err := filepath.Abs("../test/json")
	assert.NoError(t, err)
	assert.NoError(t, p.LoadFolder(folder, []string{}))

	assert.NoError(t,
		assertRoute(t, p, http.MethodPost, "http://localhost/loans/id/lock-transactions", "loans_transactions_v2_swagger", "/loans/{loanAccountId}/lock-transactions", "applyLock", "loantransactions"),
	)
}

func TestCommon(t *testing.T) {
	p := NewURLParser()

	// load specs
	if err := p.AddSpec("users-spec", absPath("../test/json/users_v2_swagger.json"), []string{}); err != nil {
		t.Error(err)
	}
	if err := p.AddSpec("clients-spec", absPath("../test/json/clients_v2_swagger.json"), []string{}); err != nil {
		t.Error(err)
	}
	if err := p.AddSpec("loans-spec", absPath("../test/json/loans_transactions_v2_swagger.json"), []string{}); err != nil {
		t.Error(err)
	}

	// returs 404 for non-existing route in specs
	assert.EqualError(t,
		assertRoute(t, p, http.MethodGet, "http://api.mambu.com", "", "", "", ""),
		ErrRouteNotFound.Error(),
	)

	// resolves client route
	assert.NoError(t,
		assertRoute(t, p, http.MethodGet, "http://api.mambu.com/clients/id", "clients-spec", "/clients/{clientId}", "getById", "clients"),
	)
	// returs 404 for the same route but another method
	assert.EqualError(t,
		assertRoute(t, p, http.MethodPost, "http://api.mambu.com/clients/id", "", "", "", ""),
		ErrRouteNotFound.Error(),
	)

	// resolves loan route
	assert.NoError(t,
		assertRoute(t, p, http.MethodPost, "http://localhost/loans/id/lock-transactions", "loans-spec", "/loans/{loanAccountId}/lock-transactions", "applyLock", "loantransactions"),
	)

	// check caching function
	assert.NoError(t,
		assertRoute(t, p, http.MethodPost, "http://localhost/loans/id/lock-transactions", "loans-spec", "/loans/{loanAccountId}/lock-transactions", "applyLock", "loantransactions"),
	)

}
