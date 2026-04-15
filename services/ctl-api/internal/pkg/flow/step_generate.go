package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	workflowsflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow"
)

// GenerateSteps generates workflow steps and persists them. It supports two modes:
// 1. GenerateStepsSignal path — enqueues the signal and calls FetchSteps update
// 2. Generator map path — uses the hardcoded generator for the workflow type
func GenerateSteps(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, generators map[app.WorkflowType]WorkflowStepGenerator) (*app.Workflow, error) {
	var steps []*app.WorkflowStep
	var err error

	if flw.GenerateStepsSignal != nil && flw.GenerateStepsSignal.Signal != nil {
		steps, err = GenerateStepsViaSignal(ctx, cfg, flw)
	} else if generators != nil {
		gen, has := generators[flw.Type]
		if !has {
			return nil, errors.Errorf("no workflow step generator registered for workflow type %s", flw.Type)
		}
		steps, err = gen(ctx, flw)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
		}
	} else {
		return nil, errors.Errorf("no step generation method available for workflow %s", flw.ID)
	}

	if err != nil {
		return nil, err
	}

	steps, err = workflowsflow.AwaitGenerateWorkflowSteps(ctx, &workflowsflow.GenerateWorkflowStepsRequest{
		WorkflowID: flw.ID,
		Steps:      steps,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create steps for workflow")
	}

	flw.Steps = make([]app.WorkflowStep, len(steps))
	for i, step := range steps {
		flw.Steps[i] = *step
	}

	return flw, nil
}

// GenerateStepsViaSignal enqueues the workflow's GenerateStepsSignal, then sends
// a "FetchSteps" update to the signal's handler workflow to retrieve the steps.
func GenerateStepsViaSignal(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	// Set the workflow ID on the signal so its Execute method can look up the workflow.
	type workflowIDSetter interface {
		SetWorkflowID(id string)
	}
	if setter, ok := flw.GenerateStepsSignal.Signal.(workflowIDSetter); ok {
		setter.SetWorkflowID(flw.ID)
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         cfg.OwnerID,
		OwnerType:       cfg.OwnerType,
		QueueName:       cfg.QueueName,
		Signal:          flw.GenerateStepsSignal.Signal,
		SignalOwnerID:   flw.ID,
		SignalOwnerType: "install_workflows",
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue generate-steps signal")
	}

	steps, err := queueclient.AwaitFetchSteps(ctx, queueclient.FetchStepsRequest{
		QueueSignalID: enqueueResp.QueueSignalID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch steps from generate-steps signal")
	}

	return steps, nil
}
