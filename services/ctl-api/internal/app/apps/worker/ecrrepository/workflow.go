package ecrrepository

import (
	workers "github.com/nuonco/nuon/services/ctl-api/internal"
)

type Wkflow struct {
	Cfg *workers.Config
}

func NewWorkflow(cfg *workers.Config) Wkflow {
	return Wkflow{
		Cfg: cfg,
	}
}
