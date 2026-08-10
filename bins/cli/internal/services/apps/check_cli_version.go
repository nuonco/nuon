package apps

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// warnIfCLIOutdated warns when this CLI predates the control plane's floor for
// server-side app config sync. Below that floor the CLI syncs app configs itself and
// never sends action or runbook ids, so new ones never reach installs and the sync still
// reports success. Never fatal — an unreachable or older control plane means no warning.
func (s *Service) warnIfCLIOutdated(ctx context.Context) {
	if version.IsDev() {
		return
	}

	current, err := semver.NewVersion(version.Version)
	if err != nil {
		return
	}

	cp := version.FetchControlPlane(ctx, s.cfg.APIURL)
	if cp == nil || cp.MinCLIForServerSideSync() == "" {
		return
	}

	minimum, err := semver.NewVersion(cp.MinCLIForServerSideSync())
	if err != nil || !current.LessThan(minimum) {
		return
	}

	ui.PrintWarning(fmt.Sprintf(
		"your CLI (%s) is older than the minimum %s this control plane syncs app configs for — actions and runbooks will not be synced to installs. see https://docs.nuon.co/cli to update.",
		current, minimum,
	))
}
