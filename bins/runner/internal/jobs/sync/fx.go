package sync

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	noop "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/noop"
	oci "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/oci"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sync", oci.New)),
		fx.Provide(jobs.AsJobHandler("sync", noop.New)),
	}
}
