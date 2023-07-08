package gqlclient

import (
	"net/http"
)

// authTransport is a transport that injects our authentication token into the api
type authTransport struct {
	authToken string
	transport http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.authToken)
	return t.transport.RoundTrip(req)
}
