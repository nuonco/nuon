package check

import (
	"time"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

// NOTE(fd): borrowed from noop
type HealthcheckConfig configs.HealthcheckConfig

type handlerState struct {
	// set during the fetch/validate phase
	cfg     *HealthcheckConfig
	timeout time.Duration

	// fields set by the plugin execution
	jobExecutionID string
	jobID          string
	outputs        configs.HealthcheckOutputs
}
