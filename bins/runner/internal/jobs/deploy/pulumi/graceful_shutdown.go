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

	// Unlike Terraform, Pulumi doesn't have a separate lock mechanism to clean up.
	// The Automation API handles its own cleanup on cancellation.
	l.Info("pulumi graceful shutdown - no additional cleanup needed")

	return nil
}
