package terraform

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

type (
	Build          configs.Build[configs.NoopBuild, configs.Registry[configs.NoopRegistry]]
	Deploy         configs.Deploy[configs.RunnerTerraform]
	WaypointConfig configs.Apps[Build, Deploy]
)

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.RunnerTerraform
	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
}
