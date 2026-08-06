package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// warnIfCLIOutdated warns when this CLI predates the control plane's recommended
// version. Below that floor the CLI syncs app configs itself and never sends action or
// runbook ids, so a sync silently drops them from installs and still reports success.
// Never fatal — an unreachable or older control plane just means no warning.
func (s *Service) warnIfCLIOutdated(ctx context.Context) {
	if version.Version == "development" {
		return
	}

	current, err := semver.NewVersion(version.Version)
	if err != nil {
		return
	}

	recommended := s.recommendedCLIVersion(ctx)
	if recommended == nil || !current.LessThan(recommended) {
		return
	}

	ui.PrintWarning(fmt.Sprintf(
		"your CLI (%s) is older than the recommended %s — actions and runbooks will not be synced to installs. see https://docs.nuon.co/cli to update.",
		current, recommended,
	))
}

func (s *Service) recommendedCLIVersion(ctx context.Context) *semver.Version {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.APIURL+"/version", nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var body struct {
		RecommendedCLIVersion string `json:"recommended_cli_version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.RecommendedCLIVersion == "" {
		return nil
	}

	recommended, err := semver.NewVersion(body.RecommendedCLIVersion)
	if err != nil {
		return nil
	}
	return recommended
}
