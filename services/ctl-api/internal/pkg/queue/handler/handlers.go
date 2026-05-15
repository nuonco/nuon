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
