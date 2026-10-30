package publicnet

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

type httpTransport struct {
	transport *http.Transport
}

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := validateHTTPURL(req.URL); err != nil {
		return nil, err
	}
	return t.transport.RoundTrip(req)
}

func (t *httpTransport) CloseIdleConnections() {
	t.transport.CloseIdleConnections()
}

func validateHTTPURL(u *url.URL) error {
	if (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" || u.Hostname() == "" || u.User != nil {
		return fmt.Errorf("source must be an HTTP(S) URL without credentials")
	}
	return nil
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: &httpTransport{transport: &http.Transport{
			DialContext:           DialContext(&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}),
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			ForceAttemptHTTP2:     true,
		}},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many source redirects")
			}
			req.Header.Del("Referer")
			if err := validateHTTPURL(req.URL); err != nil {
				return err
			}
			return RequirePublicHost(req.Context(), req.URL.Hostname())
		},
	}
}
