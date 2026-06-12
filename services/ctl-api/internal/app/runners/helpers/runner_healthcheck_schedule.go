package helpers

import (
	"time"

	config "github.com/nuonco/nuon/pkg/services/config"
)

const runnerHealthcheckSignalExpiry = 5 * time.Minute

func runnerHealthcheckSchedule(env config.Env) (schedule string) {
	if env == config.Development {
		return "* * * * *"
	}
	return "*/15 * * * *"
}
