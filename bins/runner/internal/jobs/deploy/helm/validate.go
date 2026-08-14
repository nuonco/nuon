package helm

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	ociarchive "github.com/nuonco/nuon/pkg/runner/oci/archive"
)

const (
	defaultNamespace string = "default"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// A recovery rebuilds the target release from the revision helm already
	// stored, so it needs no chart. Skipping the artifact is not just an
	// optimisation: requiring it would make recovery fail whenever the artifact
	// is unreachable, which is exactly when an install is most likely wedged.
	if h.isRecovery() {
		l.Info("recovery job, skipping chart artifact")
		return nil
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
