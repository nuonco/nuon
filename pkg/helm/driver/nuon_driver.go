package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	rspb "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var _ driver.Driver = (*Nuon)(nil)

const (
	// NuonDriverName is the string name of this driver.
	NuonDriverName = "nuon"
)

// Nuon is the Nuon storage driver implementation.
type Nuon struct {
	client    *http.Client
	serverURL string
	namespace string
	headers   map[string]string
}

// NewNuonDriver initializes a new Nuon driver.
func NewNuonDriver(serverURL, apiKey string) (*Nuon, error) { // Accept headers as a parameter
	if _, err := url.Parse(serverURL); err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	return &Nuon{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		serverURL: strings.TrimSuffix(serverURL, "/"),
		namespace: "default",
		headers: map[string]string{
			"Authorization": "Bearer " + apiKey,
		},
	}, nil
}

// SetNamespace sets a specific namespace in which releases will be accessed.
func (h *Nuon) SetNamespace(ns string) {
	h.namespace = ns
}

// Name returns the name of the driver.
func (h *Nuon) Name() string {
	return NuonDriverName
}

// Get returns the release named by key or returns ErrReleaseNotFound.
func (h *Nuon) Get(key string) (*rspb.Release, error) {
	endpoint := fmt.Sprintf("%s/releases/%s/%s", h.serverURL, h.namespace, url.PathEscape(key))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, driver.ErrReleaseNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get release: %s", resp.Status)
	}

	var release rspb.Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// List returns the list of all releases such that filter(release) == true
func (h *Nuon) List(filter func(*rspb.Release) bool) ([]*rspb.Release, error) {
	endpoint := fmt.Sprintf("%s/releases/%s", h.serverURL, h.namespace)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list releases: %s", resp.Status)
	}

	var releases []*rspb.Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	var filtered []*rspb.Release
	for _, rls := range releases {
		if filter(rls) {
			filtered = append(filtered, rls)
		}
	}

	return filtered, nil
}

// Query returns the set of releases that match the provided set of labels
func (h *Nuon) Query(keyvals map[string]string) ([]*rspb.Release, error) {
	params := url.Values{}
	for k, v := range keyvals {
		params.Add(k, v)
	}

	endpoint := fmt.Sprintf("%s/releases/%s/query?%s", h.serverURL, h.namespace, params.Encode())

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, driver.ErrReleaseNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query releases: %s", resp.Status)
	}

	var releases []*rspb.Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// Create creates a new release or returns ErrReleaseExists.
func (h *Nuon) Create(key string, rls *rspb.Release) error {
	// For backwards compatibility, we protect against an unset namespace
	namespace := rls.Namespace
	if namespace == "" {
		namespace = "default"
	}
	h.SetNamespace(namespace)

	data, err := json.Marshal(rls)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/releases/%s/%s", h.serverURL, h.namespace, url.PathEscape(key))

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return driver.ErrReleaseExists
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create release: %s, %s", resp.Status, body)
	}

	return nil
}

// Update updates a release or returns ErrReleaseNotFound.
func (h *Nuon) Update(key string, rls *rspb.Release) error {
	// For backwards compatibility, we protect against an unset namespace
	namespace := rls.Namespace
	if namespace == "" {
		namespace = "default"
	}
	h.SetNamespace(namespace)

	data, err := json.Marshal(rls)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/releases/%s/%s", h.serverURL, h.namespace, url.PathEscape(key))

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return driver.ErrReleaseNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update release: %s, %s", resp.Status, body)
	}

	return nil
}

// Delete deletes a release or returns ErrReleaseNotFound.
func (h *Nuon) Delete(key string) (*rspb.Release, error) {
	release, err := h.Get(key)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/releases/%s/%s", h.serverURL, h.namespace, url.PathEscape(key))

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, driver.ErrReleaseNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to delete release: %s, %s", resp.Status, body)
	}

	return release, nil
}
