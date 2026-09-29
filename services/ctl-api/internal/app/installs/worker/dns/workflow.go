package installdelegationdns

import (
	workers "github.com/nuonco/nuon/services/ctl-api/internal"
)

func NewWorkflow(cfg workers.Config) Wkflow {
	return Wkflow{
		cfg: cfg,
	}
}

type Wkflow struct {
	cfg workers.Config
}
