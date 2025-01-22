package terraform

import (
	"time"

	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	terraformworkspace "github.com/powertoolsdev/mono/pkg/terraform/workspace"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	defaultFileType string = "file/terraform"
)

type (
	Build          configs.NoRegistryBuild[configs.DockerRefBuild]
	Deploy         configs.Deploy[configs.TerraformDeploy]
	WaypointConfig configs.Apps[Build, Deploy]
)

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.TerraformDeploy
	srcCfg  *configs.OCIRegistryRepository
	srcTag  string
	timeout time.Duration

	// fields set by the plugin execution
	arch           ociarchive.Archive
	jobExecutionID string
	jobID          string
	tfWorkspace    terraformworkspace.Workspace
}
