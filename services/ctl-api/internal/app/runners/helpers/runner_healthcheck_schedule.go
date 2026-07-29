package helpers

import (
	"time"

	config "github.com/nuonco/nuon/pkg/services/config"
)

const runnerHealthcheckSignalExpiry = 5 * time.Minute

// runnerHealthcheckJitterWindow spreads healthcheck emitters across the full
// prod schedule interval by shifting each emitter's cron minute field.
const runnerHealthcheckJitterWindow = 15 * time.Minute

func runnerHealthcheckSchedule(env config.Env) (schedule string) {
	if env == config.Development {
		return "* * * * *"
	}
	return "*/15 * * * *"
}
