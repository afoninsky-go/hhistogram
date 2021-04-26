package openapi

import (
	"net/http"
	"net/url"
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

func TestCommon(t *testing.T) {
	p := NewURLParser()

	// load specs
	if err := p.AddSpec("users-spec", "./users.json", []string{}); err != nil {
		t.Error(err)
	}
	if err := p.AddSpec("clients-spec", "./clients.json", []string{}); err != nil {
		t.Error(err)
	}
	if err := p.AddSpec("loans-spec", "./loans.json", []string{}); err != nil {
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

}
