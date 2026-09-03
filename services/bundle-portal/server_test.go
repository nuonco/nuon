package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestPortalContentSecurityPolicyAllowsRuntimeStyles(t *testing.T) {
	portal, err := newPortalServer("csrf", map[string]bool{"portal.test": true}, zaptest.NewLogger(t))
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "http://portal.test/", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(
		t,
		"default-src 'self'; connect-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; frame-ancestors 'none'",
		response.Header().Get("Content-Security-Policy"),
	)
}
