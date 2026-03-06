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

// ListWorkflowFns returns the list of workflow functions for registration
func (w *Wkflow) ListWorkflowFns() []any {
	return []any{
		w.ProvisionECRRepository,
		w.DeprovisionECRRepository,
	}
}
