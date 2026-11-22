package actions

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	workflow "github.com/nuonco/nuon/pkg/runner/jobs/actions/workflow"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("actions", workflow.New)),
	}
}
