package installdelegationdns

import (
	workers "github.com/powertoolsdev/mono/services/ctl-api/internal"
)

const (
	defaultNuonRunDomain string = "nuon.run"
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
