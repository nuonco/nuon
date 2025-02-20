package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/actions"
	actionsworkflow "github.com/powertoolsdev/mono/bins/runner/internal/jobs/actions/workflow"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/build"
	containerimagebuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/containerimage"
	dockerbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/docker"
	helmbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/helm"
	noopbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/noop"
	terraformbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy"
	helmdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/helm"
	jobdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/job"
	noopdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/noop"
	terraformdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/healthcheck"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/runner"
	runnerhelm "github.com/powertoolsdev/mono/bins/runner/internal/jobs/runner/helm"
	runnerterraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/runner/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox"
	sandboxterraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync"
	"github.com/powertoolsdev/mono/bins/runner/internal/registry"

	// containerimagesync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/containerimage"
	noopsync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/noop"
	ocisync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/oci"

	noopoperation "github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations/noop"
	shutdownoperation "github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations/shutdown"
	updateoperation "github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations/update"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
)

func (c *cli) registerRun() error {
	runCmd := &cobra.Command{
		Use:  "run",
		Long: "run executes the runner job loop, and runs and manages jobs until interrupted.",
		Run:  c.runRun,
	}

	rootCmd.AddCommand(runCmd)
	return nil
}

func (c *cli) runRun(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		// build jobs
		fx.Provide(jobloop.AsJobLoop(build.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("builds", dockerbuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", containerimagebuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", helmbuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", terraformbuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", noopbuild.New)),

		// deploy jobs
		fx.Provide(jobloop.AsJobLoop(deploy.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("deploys", helmdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", jobdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", noopdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", terraformdeploy.New)),

		// sync jobs
		fx.Provide(jobloop.AsJobLoop(sync.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sync", ocisync.New)),
		fx.Provide(jobs.AsJobHandler("sync", noopsync.New)),

		// healthcheck jobs
		fx.Provide(jobloop.AsJobLoop(healthcheck.NewJobLoop)),

		// operation jobs
		fx.Provide(jobloop.AsJobLoop(operations.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("operations", noopoperation.New)),
		fx.Provide(jobs.AsJobHandler("operations", shutdownoperation.New)),
		fx.Provide(jobs.AsJobHandler("operations", updateoperation.New)),

		// sandbox jobs
		fx.Provide(jobloop.AsJobLoop(sandbox.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sandbox", sandboxterraform.New)),

		// runner jobs
		fx.Provide(jobloop.AsJobLoop(runner.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("runner", runnerterraform.New)),
		fx.Provide(jobs.AsJobHandler("runner", runnerhelm.New)),

		// actions jobs
		fx.Provide(jobloop.AsJobLoop(actions.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("actions", actionsworkflow.New)),

		// start all job loops
		fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
		fx.Invoke(func(*heartbeater.HeartBeater) {}),
		fx.Invoke(func(*registry.Registry) {}),
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
