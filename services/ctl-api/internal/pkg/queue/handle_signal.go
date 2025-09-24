package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
	handleractivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler/activities"
)

func (q *queue) handleQueueSignal(ctx workflow.Context, queueRef QueueRef) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("starting processing of queue signal")
	queueSignal, err := activities.AwaitGetQueueSignalByQueueSignalID(ctx, queueRef.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	l.Info("making sure queue signal workflow is ready")
	if _, err := handleractivities.AwaitUpdateWorkflowReady(ctx, &handleractivities.UpdateWorkflowReadyRequest{
		Namespace:  queueSignal.Workflow.Namespace,
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
	}); err != nil {
		return errors.Wrap(err, "unable to update")
	}

	l.Info("making sure queue signal workflow is valid")
	if _, err := handleractivities.AwaitUpdateWorkflowValidate(ctx, &handleractivities.UpdateWorkflowValidateRequest{
		Namespace:  queueSignal.Workflow.Namespace,
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
	}); err != nil {
		return errors.Wrap(err, "unable to validate")
	}

	if _, err := handleractivities.AwaitUpdateWorkflowExecute(ctx, &handleractivities.UpdateWorkflowExecuteRequest{
		Namespace:  queueSignal.Workflow.Namespace,
		UpdateID:   queueSignal.ID,
		WorkflowID: queueRef.WorkflowID,
	}); err != nil {
		return errors.Wrap(err, "unable to validate")
	}

	// now, call the update method here
	return nil
}
