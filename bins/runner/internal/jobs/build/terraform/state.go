package terraform

import (
	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	defaultFileType string = "file/terraform"
)

type WaypointConfig configs.App[configs.Build[configs.OCIArchiveBuild, configs.OCIRegistryRepository], configs.NoopDeploy]

type handlerState struct {
	// set during the fetch/validate phase
	plan   *planv1.Plan
	cfg    *configs.OCIArchiveBuild
	dstCfg *configs.OCIRegistryRepository

	// fields set by the plugin execution
	workspace      workspace.Workspace
	arch           ociarchive.Archive
	resultTag      string
	jobExecutionID string
	jobID          string
}
