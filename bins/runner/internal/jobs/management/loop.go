package management

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

const (
	jobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupManagement
)

type SyncParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"management"`
}

func NewJobLoop(params SyncParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, jobGroup, params.BaseParams)
}
