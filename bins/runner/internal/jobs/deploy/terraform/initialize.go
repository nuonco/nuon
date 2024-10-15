package terraform

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing archive...")
	if err := h.state.arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}

	l.Info("unpacking archive...")
	if err := h.state.arch.Unpack(ctx, h.state.srcCfg, h.state.srcTag); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	return nil
}
