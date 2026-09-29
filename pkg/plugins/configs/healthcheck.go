package configs

import "time"

type HealthcheckConfig struct {
	Noop bool
}

type HealthcheckOutputs struct {
	JobLoops map[string]time.Duration `json:"job_loops"`
}
