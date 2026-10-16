package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

const ResumeRunnerJobUpdateName = "resume-runner-job"

func (h *handler) resumeRunnerJob(ctx workflow.Context) (app.Status, error) {
	if h.finished {
		if isTerminalQueueStatus(h.finishedStatus) &&
			workflow.GetVersion(ctx, handlerTerminalDemandVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			return h.finishedStatus, nil
		}
		return "", errors.New("handler already finished")
	}
	if h.executing {
		return "", errors.New("handler is executing")
	}
	if h.queueSignal == nil || !h.queueSignal.RunnerDemand {
		return "", errors.New("signal does not support runner demand")
	}
	_, err := h.executeHandler(ctx, callback.Ref{})
	if h.finished {
		return h.finishedStatus, nil
	}
	return h.finishedStatus, err
}

type handlerType string

const (
	handlerTypeQuery  handlerType = "query"
	handlerTypeUpdate handlerType = "update"
)

type workflowHandler struct {
	name             string
	typ              handlerType
	handler          any
	handlerValidator any
}

func (h *handler) registerHandlers(ctx workflow.Context) error {
	handlers := []workflowHandler{
		{StatusQueryName, handlerTypeQuery, h.statusHandler, nil},
		{ReadyHandlerName, handlerTypeUpdate, h.readyHandler, nil},
		{ValidateUpdateName, handlerTypeUpdate, h.validateHandler, nil},
		{ExecuteUpdateName, handlerTypeUpdate, h.executeHandler, nil},
		{ResumeRunnerJobUpdateName, handlerTypeUpdate, h.resumeRunnerJob, nil},
		{CancelUpdateName, handlerTypeUpdate, h.cancelHandler, nil},
		{FinishedHandlerName, handlerTypeUpdate, h.finishedHandler, nil},
	}

	for _, wh := range handlers {
		switch wh.typ {
		// register query handler
		case handlerTypeQuery:
			if err := workflow.SetQueryHandler(ctx, wh.name, wh.handler); err != nil {
				return errors.Wrapf(err, "unable to create query handler %s", wh.name)
			}
			// register update handler
		case handlerTypeUpdate:
			opts := workflow.UpdateHandlerOptions{
				Validator: wh.handlerValidator,
			}
			if err := workflow.SetUpdateHandlerWithOptions(ctx, wh.name, wh.handler, opts); err != nil {
				return errors.Wrapf(err, "unable to create update handler %s", wh.name)
			}
		}
	}

	return nil
}
