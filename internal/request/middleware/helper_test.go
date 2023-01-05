package middleware

import (
	"net/http"
	"net/http/httptest"
)

// PerformRequest for testing gin router.
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
