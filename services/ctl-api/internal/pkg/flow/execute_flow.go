package flow

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowsflow "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (c *WorkflowConductor[SignalType]) Handle(ctx workflow.Context, req eventloop.EventLoopRequest, fid string) error {
	// generate steps
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, fid)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow object")
	}
	if flw.Status.Status == app.StatusCancelled {
		return errors.New("workflow already cancelled")
	}

	defer func() {
		// NOTE(jm): this should be a helper function.
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: fid,
				Status: app.CompositeStatus{
					Status: app.StatusCancelled,
				},
			}); err != nil {
				l.Error("unable to update status on cancellation", zap.Error(err))
			}
		}
	}()

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStartedAtByID(ctx, fid); err != nil {
		return err
	}
	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, fid); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	l.Debug("generating steps for workflow")
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: fid,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "generating steps for workflow",
		},
	}); err != nil {
		return err
	}

	flw, err = c.generateSteps(ctx, flw)
	if err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: fid,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "error while generating steps",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to generate workflow steps")
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: fid,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "successfully generated all steps",
		},
	}); err != nil {
		return err
	}

	l.Debug("executing steps for workflow")
	if err := c.executeSteps(ctx, req, flw); err != nil {
		status := app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: "error while executing steps",
			Metadata: map[string]any{
				"error_message": err.Error(),
			},
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     fid,
			Status: status,
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to execute workflow steps")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: fid,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "successfully executed workflow",
		},
	}); err != nil {
		return err
	}

	return nil
}

func (c *WorkflowConductor[DomainSignal]) generateSteps(ctx workflow.Context, flw *app.Workflow) (*app.Workflow, error) {
	gen, has := c.Generators[flw.Type]
	if !has {
		return nil, errors.Errorf("no workflow step generator registered for workflow type %s", flw.Type)
	}

	steps, err := gen(ctx, flw)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
	}

	steps, err = workflowsflow.AwaitGenerateWorkflowSteps(ctx, &workflowsflow.GenerateWorkflowStepsRequest{
		WorkflowID: flw.ID,
		Steps:      steps,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create steps for workflow")
	}

	// TODO(sdboyer) remove this once types align
	flw.Steps = make([]app.WorkflowStep, len(steps))
	for i, step := range steps {
		flw.Steps[i] = *step
	}

	return flw, nil
}
