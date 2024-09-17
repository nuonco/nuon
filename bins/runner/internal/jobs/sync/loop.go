package sync

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupSync
)

type SyncParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"sync"`
}

func NewJobLoop(params SyncParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, runnerJobGroup, params.BaseParams)
}
