package operations

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"

	noop "github.com/nuonco/nuon/bins/runner/internal/jobs/operations/noop"
	shutdown "github.com/nuonco/nuon/bins/runner/internal/jobs/operations/shutdown"
	update "github.com/nuonco/nuon/bins/runner/internal/jobs/operations/update"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsOperationsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("operations", noop.New)),
		fx.Provide(jobs.AsJobHandler("operations", shutdown.New)),
		fx.Provide(jobs.AsJobHandler("operations", update.New)),
	}
}
