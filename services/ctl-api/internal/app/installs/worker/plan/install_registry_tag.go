package plan

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Keyed on the build, not the deploy: per-deploy tags grow unbounded on an unchanged manifest and eventually hit the registry's per-manifest tag cap.
func installRegistryTag(deploy *app.InstallDeploy) string {
	return deploy.ComponentBuildID
}
