package configs

import "time"

type HealthcheckConfig struct {
	// we want to be able to send noop healthchecks
	Noop bool
}

type HealthcheckOutputs struct {
	JobLoops map[string]time.Duration `json:"job_loops"`
}
