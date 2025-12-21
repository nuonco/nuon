package emitter

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const (
	StopUpdateName          string = "stop"
	RestartUpdateName       string = "restart"
	StatusUpdateName        string = "status"
	EnsureRunningUpdateName string = "ensure-running"
	PauseUpdateName         string = "pause"
	ResumeUpdateName        string = "resume"
)

type handlerType string

const (
	handlerTypeUpdate handlerType = "update"
)

type handler struct {
	typ              handlerType
	handler          any
	handlerValidator any
}

func (e *emitterWorkflow) registerHandlers(ctx workflow.Context) error {
	updateHandlers := map[string]handler{
		StopUpdateName: {
			handlerTypeUpdate,
			e.stopHandler,
			nil,
		},
		RestartUpdateName: {
			handlerTypeUpdate,
			e.restartHandler,
			nil,
		},
		StatusUpdateName: {
			handlerTypeUpdate,
			e.statusHandler,
			nil,
		},
		EnsureRunningUpdateName: {
			handlerTypeUpdate,
			e.ensureRunningHandler,
			nil,
		},
		PauseUpdateName: {
			handlerTypeUpdate,
			e.pauseHandler,
			nil,
		},
		ResumeUpdateName: {
			handlerTypeUpdate,
			e.resumeHandler,
			nil,
		},
	}

	for name, h := range updateHandlers {
		opts := workflow.UpdateHandlerOptions{
			Validator: h.handlerValidator,
		}
		if err := workflow.SetUpdateHandlerWithOptions(ctx, name, h.handler, opts); err != nil {
			return errors.Wrapf(err, "unable to create update handler %s", name)
		}
	}

	return nil
}
