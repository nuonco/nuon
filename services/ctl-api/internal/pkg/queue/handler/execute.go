package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const ExecuteUpdateName string = "execute"

const executeUpdateType = handlerTypeUpdate

type ExecuteResponse struct{}

func (w *handler) executeHandler(ctx workflow.Context) (*ExecuteResponse, error) {
	defer func() {
		w.finished = true
	}()

	if err := w.sig.Execute(ctx); err != nil {
		return nil, errors.Wrap(err, "execute method failed")
	}

	return nil, nil
}
