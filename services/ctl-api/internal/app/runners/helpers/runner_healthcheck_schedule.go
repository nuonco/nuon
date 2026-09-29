package helpers

import (
	"time"

	config "github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cronutil"
)

const ProcessHealthcheckSchedule = "*/5 * * * *"

const runnerHealthcheckSignalExpiry = 5 * time.Minute

const runnerHealthcheckJitterWindow = cronutil.MaxJitterWindow

func RunnerHealthcheckSchedule(env config.Env) (schedule string) {
	if env == config.Development {
		return "* * * * *"
	}
	return "*/15 * * * *"
}
