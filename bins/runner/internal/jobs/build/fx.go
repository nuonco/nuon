package build

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	containerimagebuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/containerimage"
	docker "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/docker"
	helm "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/helm"
	noop "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/noop"
	terraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("builds", docker.New)),
		fx.Provide(jobs.AsJobHandler("builds", containerimagebuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", helm.New)),
		fx.Provide(jobs.AsJobHandler("builds", terraform.New)),
		fx.Provide(jobs.AsJobHandler("builds", noop.New)),
	}
}
