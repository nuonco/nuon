package sandbox

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
	"go.uber.org/fx"

	syncsecrets "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox/sync_secrets"
	terraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox/terraform"
)

func GetJobs() []fx.Option {
	providers := []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sandbox", terraform.New)),
		fx.Provide(jobs.AsJobHandler("sandbox", syncsecrets.New)),
	}
	return providers
}
