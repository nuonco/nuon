package auth

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

// printVersionNotice reports a CLI/control-plane version mismatch after login. Both are
// promoted from the same release stream, so a difference is normal and only worth
// mentioning once, at the point the user picks a control plane.
func (a *Service) printVersionNotice(ctx context.Context) {
	if version.IsDev() {
		return
	}

	cp := version.FetchControlPlane(ctx, a.cfg.APIURL)
	if cp == nil || cp.Version == "" || cp.Version == version.Version {
		return
	}

	ui.PrintLn(fmt.Sprintf(
		"CLI %s, control plane %s. see https://docs.nuon.co/cli to update.",
		version.Version, cp.Version,
	))
}
