package management

import (
	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/monitor"

	noop "github.com/nuonco/nuon/bins/runner/internal/jobs/management/noop"
	shutdown "github.com/nuonco/nuon/bins/runner/internal/jobs/management/shutdown"
	update "github.com/nuonco/nuon/bins/runner/internal/jobs/management/update"
	vmshutdown "github.com/nuonco/nuon/bins/runner/internal/jobs/management/vm_shutdown"
	"go.uber.org/fx"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(monitor.New),
		fx.Provide(jobs.AsJobHandler("management", noop.New)),
		fx.Provide(jobs.AsJobHandler("management", update.New)),
		fx.Provide(jobs.AsJobHandler("management", shutdown.New)),
		fx.Provide(jobs.AsJobHandler("management", vmshutdown.New)),
		fx.Provide(jobloop.AsManagementJobLoop(NewJobLoop)),
		fx.Invoke(jobloop.WithManagementJobLoops(func([]jobloop.JobLoop) {})),
		fx.Invoke(func(*monitor.Monitor) {}),
	}

}
