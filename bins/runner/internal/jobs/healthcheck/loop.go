package healthcheck

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"

	"github.com/nuonco/nuon-runner-go/models"
)

const (
	jobGroup models.AppRunnerJobGroup = models.AppRunnerJobGroupHealthDashChecks
)

type JobLoopParams struct {
	jobloop.BaseParams

	JobHandlers []jobs.JobHandler `group:"healthchecks"`
}

func NewJobLoop(params JobLoopParams) jobloop.JobLoop {
	return jobloop.New(params.JobHandlers, jobGroup, params.BaseParams)
}
