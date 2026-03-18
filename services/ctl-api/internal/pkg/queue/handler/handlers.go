package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

type handlerType string

const (
	handlerTypeQuery  handlerType = "query"
	handlerTypeUpdate handlerType = "update"
)

type workflowHandler struct {
	typ              handlerType
	handler          any
	handlerValidator any
}

func (h *handler) registerHandlers(ctx workflow.Context) error {
	handlers := map[string]workflowHandler{
		StatusQueryName: {
			handlerTypeQuery,
			h.statusHandler,
			nil,
		},
		ReadyHandlerName: {
			handlerTypeUpdate,
			h.readyHandler,
			nil,
		},
		ValidateUpdateName: {
			handlerTypeUpdate,
			h.validateHandler,
			nil,
		},
		ExecuteUpdateName: {
			handlerTypeUpdate,
			h.executeHandler,
			nil,
		},
		CancelUpdateName: {
			handlerTypeUpdate,
			h.cancelHandler,
			nil,
		},
		FinishedHandlerName: {
			handlerTypeUpdate,
			h.finishedHandler,
			nil,
		},
		SleepUpdateName: {
			handlerTypeUpdate,
			h.sleepHandler,
			nil,
		},
	}

	for name, handler := range handlers {
		switch handler.typ {
		// register query handler
		case handlerTypeQuery:
			if err := workflow.SetQueryHandler(ctx, string(name), handler.handler); err != nil {
				return errors.Wrapf(err, "unable to create query handler %s", name)
			}
			// register update handler
		case handlerTypeUpdate:
			opts := workflow.UpdateHandlerOptions{
				Validator: handler.handlerValidator,
			}
			if err := workflow.SetUpdateHandlerWithOptions(ctx, name, handler.handler, opts); err != nil {
				return errors.Wrapf(err, "unable to create update handler %s", name)
			}
		}
	}

	return nil
}
