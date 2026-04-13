package pulumi

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	ociarchive "github.com/nuonco/nuon/bins/runner/internal/pkg/oci/archive"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("initializing archive...")
	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch

	l.Info("unpacking archive...")
	if err := arch.Unpack(ctx, h.state.plan.Src, h.state.plan.SrcTag); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	return nil
}
