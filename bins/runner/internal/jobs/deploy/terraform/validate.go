package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	ociarchive "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
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
	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch

	return nil
}
