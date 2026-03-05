package handler

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

const ExecuteUpdateName string = "execute"

const executeUpdateType = handlerTypeUpdate

type ExecuteResponse struct{}

func (h *handler) executeHandler(ctx workflow.Context) (*ExecuteResponse, error) {
	defer func() {
		h.finished = true
		h.executingCtx = nil
		h.executingCancel = nil
	}()

	if h.canceled {
		return nil, errors.New("signal was canceled")
	}

	execCtx, cancel := workflow.WithCancel(ctx)
	h.executingCtx = execCtx
	h.executingCancel = cancel
	defer cancel()

	if err := h.sig.Execute(execCtx); err != nil {
		return nil, errors.Wrap(err, "execute method failed")
	}

	return nil, nil
}
