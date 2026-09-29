package version

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type ControlPlane struct {
	Version string `json:"version"`

	ServerSideSyncMinCLI string `json:"server_side_sync_min_cli_version"`

	LegacyRecommendedCLI string `json:"recommended_cli_version"`
}

func (c *ControlPlane) MinCLIForServerSideSync() string {
	if c.ServerSideSyncMinCLI != "" {
		return c.ServerSideSyncMinCLI
	}
	return c.LegacyRecommendedCLI
}

func FetchControlPlane(ctx context.Context, apiURL string) *ControlPlane {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/version", nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var cp ControlPlane
	if err := json.NewDecoder(resp.Body).Decode(&cp); err != nil {
		return nil
	}
	return &cp
}

func IsDev() bool {
	return Version == "development" || Version == ""
}
