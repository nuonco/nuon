package build

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

type Params struct {
	jobloop.BaseParams

	JobHandlers []jobs.JobHandler `group:"builds"`
}

func NewJobLoop(params Params) jobloop.JobLoop {
	return jobloop.New(params.JobHandlers,
		models.AppRunnerJobGroupBuild,
		params.BaseParams)
}
