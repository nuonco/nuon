package ecrrepository

import (
	workers "github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Wkflow struct {
	Cfg *workers.Config
}

func NewWorkflow(cfg *workers.Config) Wkflow {
	return Wkflow{
		Cfg: cfg,
	}
}
