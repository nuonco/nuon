package version

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// ControlPlane is what a control plane reports about itself and the clients it expects.
type ControlPlane struct {
	Version string `json:"version"`

	// oldest CLI the control plane will sync app configs for server-side
	ServerSideSyncMinCLI string `json:"server_side_sync_min_cli_version"`

	// pre-rename name for the field above, still emitted by older control planes
	LegacyRecommendedCLI string `json:"recommended_cli_version"`
}

// MinCLIForServerSideSync tolerates control planes that predate the rename.
func (c *ControlPlane) MinCLIForServerSideSync() string {
	if c.ServerSideSyncMinCLI != "" {
		return c.ServerSideSyncMinCLI
	}
	return c.LegacyRecommendedCLI
}

// FetchControlPlane reads the control plane's /version. Callers treat a nil result as
// "no information" — this is only ever used to inform, never to block.
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

// IsDev reports whether this is a local build, which has no comparable version.
func IsDev() bool {
	return Version == "development" || Version == ""
}
