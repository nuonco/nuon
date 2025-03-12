package check

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

// NOTE(fd): borrowed from noop
type HealthcheckConfig configs.HealthcheckConfig

type handlerState struct {
	// set during the fetch/validate phase
	plan    *planv1.Plan
	cfg     *HealthcheckConfig
	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
	outputs        configs.HealthcheckOutputs
}
