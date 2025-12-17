package operations

import (
	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"go.uber.org/fx"

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
