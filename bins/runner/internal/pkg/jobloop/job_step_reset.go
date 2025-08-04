package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/workspace"
)

func (j *jobLoop) executeResetJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	if j.isSandbox(job) {
		j.execSandboxStep(ctx)
		return nil
	}

	if err := workspace.CleanupAll(ctx); err != nil {
		return errors.Wrap(err, "unable to cleanup old workspaces")
	}

	statefulHandler, ok := handler.(jobs.StatefulJobHandler)
	if !ok {
		return nil
	}

	if err := statefulHandler.Reset(ctx); err != nil {
		return err
	}

	return nil
}
