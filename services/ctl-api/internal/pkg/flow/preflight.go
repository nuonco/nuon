package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type WorkflowPreflightResult struct {
	Findings []*compositeerrors.CompositeErrorData
}

func (r *WorkflowPreflightResult) Blocked() bool {
	if r == nil {
		return false
	}
	for _, finding := range r.Findings {
		if finding == nil {
			continue
		}
		switch finding.Severity {
		case compositeerrors.SeverityError, compositeerrors.SeverityFatal:
			return true
		}
	}
	return false
}

type WorkflowPreflight func(workflow.Context, *app.Workflow, *app.GenerateStepsResult) (*WorkflowPreflightResult, error)
