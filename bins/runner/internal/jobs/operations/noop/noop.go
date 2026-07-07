package noop

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

type Params struct {
	jobloop.BaseParams

	JobHandler []jobs.JobHandler `name:"noop"`
}

func NewJobLoop(params Params) jobloop.JobLoop {
	return jobloop.New(params.JobHandler, models.AppRunnerJobGroupOperations, params.BaseParams)
}
