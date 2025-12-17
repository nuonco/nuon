package operations

import (
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
)

const (
	jobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupOperations
)

type SyncParams struct {
	jobloop.BaseParams

	Handlers []jobs.JobHandler `group:"operations"`
}

func NewJobLoop(params SyncParams) jobloop.JobLoop {
	return jobloop.New(params.Handlers, jobGroup, params.BaseParams)
}
