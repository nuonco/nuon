package workflows

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueemitter "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow"
)

type WorkflowParams struct {
	fx.In

	JobWorkflows          *job.Workflows
	WorkflowWorkflows     *workflow.Workflows
	QueueWorkflows        *queue.Workflows
	QueueEmitterWorkflows *queueemitter.Workflows
	EnqueuerWorkflows     *enqueuer.Workflows
	HandlerWorkflows      *handler.Workflows
}

type Workflows struct {
	jobWorkflows          *job.Workflows
	workflowWorkflows     *workflow.Workflows
	queueWorkflows        *queue.Workflows
	queueemitterWorkflows *queueemitter.Workflows
	enqueuerWorkflows     *enqueuer.Workflows
	handlerWorkflows      *handler.Workflows
}

func (w *Workflows) AllWorkflows() []interface{} {
	wkflows := []interface{}{
		// jobs
		w.jobWorkflows.ExecuteJob,

		// workflows
		w.workflowWorkflows.GenerateWorkflowSteps,
		w.workflowWorkflows.WaitForApprovalResponse,
	}

	wkflows = append(wkflows, w.queueWorkflows.All()...)
	wkflows = append(wkflows, w.queueemitterWorkflows.All()...)
	wkflows = append(wkflows, w.enqueuerWorkflows.All()...)
	wkflows = append(wkflows, w.handlerWorkflows.All()...)

	return wkflows
}

func NewWorkflows(params WorkflowParams) *Workflows {
	return &Workflows{
		jobWorkflows:          params.JobWorkflows,
		workflowWorkflows:     params.WorkflowWorkflows,
		queueWorkflows:        params.QueueWorkflows,
		queueemitterWorkflows: params.QueueEmitterWorkflows,
		enqueuerWorkflows:     params.EnqueuerWorkflows,
		handlerWorkflows:      params.HandlerWorkflows,
	}
}
