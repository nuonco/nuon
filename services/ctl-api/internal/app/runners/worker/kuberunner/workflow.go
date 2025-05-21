package runner

import (
	workers "github.com/powertoolsdev/mono/services/ctl-api/internal"
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) Wkflow {
	return Wkflow{
		cfg: cfg,
	}
}

type Wkflow struct {
	cfg workers.Config
}
