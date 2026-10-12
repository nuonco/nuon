package actions

import (
	"go.uber.org/fx"

	workflow "github.com/nuonco/nuon/bins/runner/internal/jobs/actions/workflow"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobloop.AsJobLoop(NewCronJobLoop)),
		fx.Provide(jobs.AsJobHandler("actions", workflow.New)),
		fx.Provide(jobs.AsJobHandler("cron-actions", workflow.New)),
	}
}
