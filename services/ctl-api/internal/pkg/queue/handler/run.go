package handler

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (h *handler) run(ctx workflow.Context) (bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	if err := h.registerHandlers(ctx); err != nil {
		return false, err
	}

	if err := h.initializeState(ctx); err != nil {
		return false, errors.Wrap(err, "unable to initialize state")
	}

	l.Debug("handler is ready")
	h.ready = true
	if err := workflow.Await(ctx, func() bool {
		return generics.AnyTrue(h.stopped, h.restarted)
	}); err != nil {
		return false, err
	}

	return false, nil
}
