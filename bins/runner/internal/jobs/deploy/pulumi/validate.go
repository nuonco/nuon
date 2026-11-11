package pulumi

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("parsing job plan to ensure correct")
	h.state.srcCfg = h.state.plan.Src
	h.state.srcTag = h.state.plan.SrcTag

	l.Info("artifact repo", zap.Any("repo", h.state.srcCfg.Repository))

	return nil
}
