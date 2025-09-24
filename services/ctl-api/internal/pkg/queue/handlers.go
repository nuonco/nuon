package queue

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

type handlerType string

const (
	handlerTypeQuery  handlerType = "query"
	handlerTypeUpdate handlerType = "update"
)

type handler struct {
	typ              handlerType
	handler          any
	handlerValidator any
}

func (w *queue) registerHandlers(ctx workflow.Context) error {
	updateHandlers := map[string]handler{
		EnqueueUpdateName: {
			handlerTypeUpdate,
			w.enqueueHandler,
			nil,
		},
		ReadyHandlerName: {
			handlerTypeUpdate,
			w.readyHandler,
			nil,
		},
		StatusHandlerName: {
			handlerTypeUpdate,
			w.statusHandler,
			nil,
		},
	}
	for name, handler := range updateHandlers {
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
