package job

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Warn("job components are no longer supported, please use an action instead")
	return nil
}
