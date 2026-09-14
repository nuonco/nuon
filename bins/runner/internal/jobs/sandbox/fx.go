package sandbox

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"

	pulumisandbox "github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox/pulumi"
	syncsecrets "github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox/sync_secrets"
	terraform "github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox/terraform"
)

func GetJobs() []fx.Option {
	providers := []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sandbox", terraform.New)),
		fx.Provide(jobs.AsJobHandler("sandbox", syncsecrets.New)),
		fx.Provide(jobs.AsJobHandler("sandbox", pulumisandbox.New)),
	}
	return providers
}
