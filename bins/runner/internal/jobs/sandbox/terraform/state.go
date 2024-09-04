package terraform

import (
	"time"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	defaultFileType string = "file/helm"
)

type WaypointConfig configs.App[configs.NoopBuild, configs.SandboxTerraform]

type handlerState struct {
	workspace workspace.Workspace

	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.SandboxTerraform
	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
}
