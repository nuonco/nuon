package pulumi

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	if h.state == nil {
		return nil
	}

	if h.state.workspace != nil {
		l.Info("attempting to update pulumi state before shutdown")
		if err := h.updatePulumiState(ctx, h.state.workspace); err != nil {
			h.writeErrorResult(ctx, "update pulumi state", err)
		} else {
			l.Info("pulumi state updated during graceful shutdown")
		}
	}

	return nil
}
