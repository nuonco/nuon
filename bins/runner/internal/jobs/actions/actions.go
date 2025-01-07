package actions

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupActions
)

type JobLoopParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"actions"`
}

func NewJobLoop(params JobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, runnerJobGroup, params.BaseParams)
}
