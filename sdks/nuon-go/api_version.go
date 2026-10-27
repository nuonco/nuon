package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIVersion is what the control plane reports from GET /version.
type APIVersion struct {
	Version                     string `json:"version"`
	GitRef                      string `json:"git_ref"`
	ServerSideSyncMinCLIVersion string `json:"server_side_sync_min_cli_version"`
	RecommendedCLIVersion       string `json:"recommended_cli_version"`
}

// GetVersion reads the control plane's public version endpoint.
func (c *client) GetVersion(ctx context.Context) (*APIVersion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIURL+"/version", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var version APIVersion
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &version, nil
}
