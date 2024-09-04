package helm

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	defaultFileType string = "file/helm"
)

type WaypointConfig configs.App[configs.NoopBuild, configs.RunnerHelm]

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *configs.RunnerHelm
	timeout time.Duration

	// fields set by the plugin execution
	chartPath      string
	jobExecutionID string
	jobID          string
}
