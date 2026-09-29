package check

import (
	"time"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

type HealthcheckConfig configs.HealthcheckConfig

type handlerState struct {
	cfg     *HealthcheckConfig
	timeout time.Duration

	jobExecutionID string
	jobID          string
	outputs        configs.HealthcheckOutputs
}
