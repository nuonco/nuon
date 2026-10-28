package http

import (
	"context"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"net/url"
	"strings"
	"time"
)

// startProxy listens on localhost and forwards only this workspace's terraform
// backend calls, attaching the runner token there. backend.json and the
// terraform state then hold the proxy address, so vendor terraform (including
// local-exec) never sees the token.
func (s *http) startProxy() error {
	if s.Config == nil || s.Config.Token == "" || s.localBase != "" {
		return nil
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("unable to listen for terraform backend proxy: %w", err)
	}

	srv := &nethttp.Server{Handler: nethttp.HandlerFunc(s.forward)}
	s.srv = srv
	go func() {
		_ = srv.Serve(ln)
	}()
	s.localBase = "http://" + ln.Addr().String()
	return nil
}

func (s *http) Close() error {
	if s.srv == nil {
		return nil
	}
	srv := s.srv
	s.srv = nil
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func (s *http) forward(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !allowedBackendPath(r.URL.Path, s.Config.WorkspaceID) {
		nethttp.NotFound(w, r)
		return
	}

	target, err := upstreamURL(s.Config.APIEndpoint, r.URL.Path, withToken(r.URL.RawQuery, s.Config.Token))
	if err != nil {
		nethttp.Error(w, "bad gateway", nethttp.StatusBadGateway)
		return
	}

	req, err := nethttp.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		nethttp.Error(w, "bad gateway", nethttp.StatusBadGateway)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Del("Host")

	resp, err := s.client().Do(req)
	if err != nil {
		nethttp.Error(w, "bad gateway", nethttp.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *http) client() *nethttp.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return nethttp.DefaultClient
}

func allowedBackendPath(path, workspaceID string) bool {
	switch path {
	case "/v1/terraform-backend":
		return true
	case "/v1/terraform-workspaces/" + workspaceID + "/lock",
		"/v1/terraform-workspaces/" + workspaceID + "/unlock":
		return workspaceID != ""
	default:
		return false
	}
}

func withToken(rawQuery, token string) string {
	q, _ := url.ParseQuery(rawQuery)
	q.Set("token", token)
	return q.Encode()
}

func backendQuery(pairs ...string) string {
	q := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		q.Set(pairs[i], pairs[i+1])
	}
	return q.Encode()
}

func upstreamURL(endpoint, path, rawQuery string) (string, error) {
	u, err := url.Parse(strings.TrimRight(endpoint, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid runner api url %q", endpoint)
	}
	u.Path = path
	u.RawQuery = rawQuery
	u.Fragment = ""
	return u.String(), nil
}
