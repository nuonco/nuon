package workflows

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

type WorkflowParams struct {
	fx.In

	JobWorkflows *job.Workflows
}

type Workflows struct {
	jobWorkflows *job.Workflows
}

func (w *Workflows) AllWorkflows() []interface{} {
	return []interface{}{
		w.jobWorkflows.ExecuteJob,
	}
}

func NewWorkflows(params WorkflowParams) *Workflows {
	return &Workflows{
		jobWorkflows: params.JobWorkflows,
	}
}
