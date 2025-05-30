package terraform

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"go.uber.org/zap"
)

func (p *handler) GracefulShutdown(ctx context.Context, job *models.AppRunnerJob, l *zap.Logger) error {
	hlog := log.NewHClog(l)
	if p.state == nil {
		return nil
	}
	if p.state.tfWorkspace != nil {
		l.Info("attempting to update terraform state before shutdown")
		err := p.updateTerraformState(ctx, p.state.tfWorkspace, hlog)
		if err != nil {
			p.writeErrorResult(ctx, "update terraform state", err)
			// we don't return an error here because we want to allow the graceful shutdown to complete even when state update fails
		}
		l.Info("terraform state updated during  graceful shutdown")
	}

	if p.state.plan != nil && p.state.plan.TerraformBackend != nil {
		l.Info("attempting to unlock terraform workspace before shutdown")
		err := p.apiClient.UnlockTerraformWorkspace(ctx, p.state.plan.TerraformBackend.WorkspaceID)
		if err != nil {
			p.writeErrorResult(ctx, "unlock terraform workspace", err)
			// we don't return an error here because we want to allow the graceful shutdown to complete even when unlock fails
		}
		l.Info("terraform workspace unlocked during graceful shutdown")
	}

	return nil
}
