package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type connectedClient struct {
	baseURL    *url.URL
	orgID      string
	installID  string
	apiToken   string
	httpClient *http.Client
}

func newConnectedClient(controlPlaneURL, orgID, installID, apiToken string) (*connectedClient, error) {
	baseURL, err := url.Parse(controlPlaneURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid --control-plane-url %q", controlPlaneURL)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("--control-plane-url must use http or https")
	}
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	return &connectedClient{
		baseURL: baseURL, orgID: orgID, installID: installID, apiToken: apiToken,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *connectedClient) proxy(w http.ResponseWriter, r *http.Request, suffix string) {
	target := *c.baseURL
	target.Path = path.Join(strings.TrimSuffix(c.baseURL.Path, "/"), "/v1/customer-managed/installs/", c.installID, suffix)
	target.RawQuery = r.URL.RawQuery
	request, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	request.Header.Set("Authorization", "Bearer "+c.apiToken)
	request.Header.Set("X-Nuon-Org-ID", c.orgID)
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		writeAPIError(w, fmt.Errorf("control plane request failed: %w", err), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	for _, header := range []string{"Content-Type", "Content-Encoding", "Cache-Control"} {
		if value := response.Header.Get(header); value != "" {
			w.Header().Set(header, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}
