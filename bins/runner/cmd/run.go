package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	containerimagebuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/containerimage"
	dockerbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/docker"
	helmbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/helm"
	noopbuild "github.com/powertoolsdev/mono/bins/runner/internal/jobs/build/noop"
	helmdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/helm"
	jobdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/job"
	noopdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/noop"
	terraformdeploy "github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations"
	runnerhelm "github.com/powertoolsdev/mono/bins/runner/internal/jobs/runner/helm"
	runnerterraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/runner/terraform"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox"
	sandboxterraform "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox/terraform"

	// containerimagesync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/containerimage"
	noopsync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/noop"
	ocisync "github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync/oci"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/user"

	noopoperation "github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations/noop"
	shutdownoperation "github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations/shutdown"
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
		// fx.Provide(jobloop.AsJobLoop(build.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("builds", dockerbuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", containerimagebuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", helmbuild.New)),
		fx.Provide(jobs.AsJobHandler("builds", noopbuild.New)),

		// deploy jobs
		// fx.Provide(jobloop.AsJobLoop(deploy.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("deploys", helmdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", jobdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", noopdeploy.New)),
		fx.Provide(jobs.AsJobHandler("deploys", terraformdeploy.New)),

		// sync jobs
		// fx.Provide(jobloop.AsJobLoop(sync.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sync", ocisync.New)),
		fx.Provide(jobs.AsJobHandler("sync", noopsync.New)),

		// healthcheck jobs
		// fx.Provide(jobloop.AsJobLoop(healthcheck.NewJobLoop)),

		// operation jobs
		fx.Provide(jobloop.AsJobLoop(operations.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("operations", noopoperation.New)),
		fx.Provide(jobs.AsJobHandler("operations", shutdownoperation.New)),

		// sandbox jobs
		fx.Provide(jobloop.AsJobLoop(sandbox.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("sandbox", sandboxterraform.New)),

		// runner jobs
		// fx.Provide(jobloop.AsJobLoop(runner.NewJobLoop)),
		fx.Provide(jobs.AsJobHandler("runner", runnerterraform.New)),
		fx.Provide(jobs.AsJobHandler("runner", runnerhelm.New)),

		// TODO(jm): add diagnostics jobs here
		// user jobs
		fx.Provide(jobloop.AsJobLoop(user.NewJobLoop)),

		// start all job loops
		fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
		//fx.Invoke(func(*heartbeater.HeartBeater) {}),
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
