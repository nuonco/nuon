package terraform

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

type WaypointConfig configs.App[configs.NoopBuild, configs.RunnerTerraform]

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.RunnerTerraform
	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
}
