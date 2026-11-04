package pulumi

import (
	"context"
	"errors"
	"fmt"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) (retErr error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	arch := ociarchive.New()
	if err := arch.Initialize(ctx); err != nil {
		return fmt.Errorf("unable to initialize archive: %w", err)
	}
	h.state.arch = arch
	defer func() {
		if retErr != nil {
			retErr = errors.Join(retErr, arch.Cleanup(context.WithoutCancel(ctx)))
			h.state.arch = nil
		}
	}()

	l.Info("unpacking archive...")
	if err := h.state.arch.Unpack(ctx, h.state.srcCfg, h.state.srcTag); err != nil {
		return fmt.Errorf("unable to unpack archive: %w", err)
	}

	return nil
}
