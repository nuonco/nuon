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
	name             string
	typ              handlerType
	handler          any
	handlerValidator any
}

func (w *queue) registerHandlers(ctx workflow.Context) error {
	// Using a slice for temporal determinism.
	updateHandlers := []handler{
		{EnqueueUpdateName, handlerTypeUpdate, w.enqueueHandler, nil},
		{ReadyHandlerName, handlerTypeQuery, w.readyHandler, nil},
		{StatusHandlerName, handlerTypeUpdate, w.statusHandler, nil},
		{PauseHandlerName, handlerTypeUpdate, w.pauseHandler, nil},
		{ResumeHandlerName, handlerTypeUpdate, w.resumeHandler, nil},
		{StopUpdateName, handlerTypeUpdate, w.stopUpdateHandler, nil},
		{DirectExecuteUpdateName, handlerTypeUpdate, w.directExecuteHandler, nil},
		{CheckCANUpdateName, handlerTypeUpdate, w.checkCANHandler, nil},
	}
	for _, h := range updateHandlers {
		switch h.typ {
		// register query handler
		case handlerTypeQuery:
			if err := workflow.SetQueryHandler(ctx, h.name, h.handler); err != nil {
				return errors.Wrapf(err, "unable to create query handler %s", h.name)
			}

			// register update handler
		case handlerTypeUpdate:
			opts := workflow.UpdateHandlerOptions{
				Validator: h.handlerValidator,
			}
			if err := workflow.SetUpdateHandlerWithOptions(ctx, h.name, h.handler, opts); err != nil {
				return errors.Wrapf(err, "unable to create update handler %s", h.name)
			}
		}
	}

	return nil
}
