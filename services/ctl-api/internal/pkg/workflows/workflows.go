package workflows

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow"
)

type WorkflowParams struct {
	fx.In

	JobWorkflows      *job.Workflows
	WorkflowWorkflows *workflow.Workflows
}

type Workflows struct {
	jobWorkflows      *job.Workflows
	workflowWorkflows *workflow.Workflows
}

func (w *Workflows) AllWorkflows() []interface{} {
	return []interface{}{
		w.jobWorkflows.ExecuteJob,
		w.workflowWorkflows.GenerateWorkflowSteps,
	}
}

func NewWorkflows(params WorkflowParams) *Workflows {
	return &Workflows{
		jobWorkflows:      params.JobWorkflows,
		workflowWorkflows: params.WorkflowWorkflows,
	}
}
