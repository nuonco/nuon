package deploy

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupDeploy
)

type JobLoopParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `name:"deploy"`
}

func NewJobLoop(params JobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, runnerJobGroup, params.BaseParams)
}
