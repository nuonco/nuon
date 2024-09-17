package user

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

const (
	runnerJobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupUser
)

type JobLoopParams struct {
	jobloop.BaseParams

	// Handlers []jobs.JobHandler `name:"user"`
}

func NewJobLoop(params JobLoopParams) jobloop.JobLoop {
	return jobloop.New(nil,
		runnerJobGroup,
		params.BaseParams)
}
