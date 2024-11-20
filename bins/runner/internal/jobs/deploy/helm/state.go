package helm

import (
	"time"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	defaultFileType string = "file/helm"
)

type (
	Build          configs.Build[configs.NoopBuild, configs.Registry[configs.NoopRegistry]]
	Deploy         configs.Deploy[configs.HelmRepoDeploy]
	WaypointConfig configs.Apps[Build, Deploy]
)

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.HelmRepoDeploy
	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	// fields set by the plugin execution
	arch           ociarchive.Archive
	chartPath      string
	jobExecutionID string
	jobID          string
}
