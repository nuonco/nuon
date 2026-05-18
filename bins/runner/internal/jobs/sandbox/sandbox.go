package sandbox

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupSandbox
)

type JobLoopParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"sandbox"`
}

func NewJobLoop(params JobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, runnerJobGroup, params.BaseParams)
}
