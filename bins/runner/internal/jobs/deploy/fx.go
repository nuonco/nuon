package deploy

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	helm "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/helm"
	job "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/job"
	kubernetesmanifest "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/kubernetes_manifest"
	noop "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/noop"
	terraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("deploys", helm.New)),
		fx.Provide(jobs.AsJobHandler("deploys", job.New)),
		fx.Provide(jobs.AsJobHandler("deploys", noop.New)),
		fx.Provide(jobs.AsJobHandler("deploys", terraform.New)),
		fx.Provide(jobs.AsJobHandler("deploys", kubernetesmanifest.New)),
	}
}
