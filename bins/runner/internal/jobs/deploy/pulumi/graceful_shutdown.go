package pulumi

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
)

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	if h.state == nil {
		return nil
	}

	if h.state.workspace != nil {
		l.Info("attempting to update pulumi state before shutdown")
		if err := h.updatePulumiState(ctx, h.state.workspace); err != nil {
			h.writeErrorResult(ctx, "update pulumi state", err)
			// Don't return the error -- allow graceful shutdown to complete
			// even when state update fails.
		} else {
			l.Info("pulumi state updated during graceful shutdown")
		}
	}

	return nil
}
