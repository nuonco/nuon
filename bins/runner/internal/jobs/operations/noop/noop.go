package noop

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

type Params struct {
	jobloop.BaseParams

	JobHandler []jobs.JobHandler `name:"noop"`
}

func NewJobLoop(params Params) jobloop.JobLoop {
	return jobloop.New(params.JobHandler, models.AppRunnerJobGroupOperations, params.BaseParams)
}
