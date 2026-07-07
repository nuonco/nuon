package sync

import (
	"go.uber.org/fx"

	noop "github.com/nuonco/nuon/bins/runner/internal/jobs/sync/noop"
	oci "github.com/nuonco/nuon/bins/runner/internal/jobs/sync/oci"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/jobs/sync/imagemetadata"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sync", oci.New)),
		fx.Provide(jobs.AsJobHandler("sync", noop.New)),
		fx.Provide(jobs.AsJobHandler("sync", imagemetadata.New)),
	}
}
