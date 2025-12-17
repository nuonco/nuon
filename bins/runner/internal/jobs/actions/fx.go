package actions

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	workflow "github.com/nuonco/nuon/bins/runner/internal/jobs/actions/workflow"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("actions", workflow.New)),
	}
}
