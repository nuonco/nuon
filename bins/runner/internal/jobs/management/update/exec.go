package update

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	// As with the shutdown job handler, fx shutdown cannot be safely triggered in this phase.
	// Must be done in cleanup.
	l.Info("exec", zap.String("job_type", "update-version"), zap.String("expected_version", h.state.expectedVersion))
	return nil
}
