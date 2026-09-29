package plan

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func installRegistryTag(deploy *app.InstallDeploy) string {
	return deploy.ComponentBuildID
}
