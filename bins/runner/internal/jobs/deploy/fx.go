package deploy

import (
	"go.uber.org/fx"

	job "github.com/nuonco/nuon/bins/runner/internal/jobs/deploy/job"
	noop "github.com/nuonco/nuon/bins/runner/internal/jobs/deploy/noop"
	pulumideploy "github.com/nuonco/nuon/bins/runner/internal/jobs/deploy/pulumi"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	helm "github.com/nuonco/nuon/pkg/runner/jobs/deploy/helm"
	kubernetesmanifest "github.com/nuonco/nuon/pkg/runner/jobs/deploy/kubernetes_manifest"
	terraform "github.com/nuonco/nuon/pkg/runner/jobs/deploy/terraform"
)

func GetJobs() []fx.Option {
	return []fx.Option{
		fx.Provide(jobloop.AsJobLoop(NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("deploys", helm.New)),
		fx.Provide(jobs.AsJobHandler("deploys", job.New)),
		fx.Provide(jobs.AsJobHandler("deploys", noop.New)),
		fx.Provide(jobs.AsJobHandler("deploys", terraform.New)),
		fx.Provide(jobs.AsJobHandler("deploys", pulumideploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", kubernetesmanifest.New)),
	}
}
