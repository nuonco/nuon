package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/actions"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/build"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync"

	"github.com/powertoolsdev/mono/bins/runner/internal/registry"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"

	check "github.com/powertoolsdev/mono/bins/runner/internal/jobs/healthcheck/check"
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
	providers := []fx.Option{}

	// common providers
	providers = append(providers, c.providers()...)

	// sandbox
	providers = append(providers, sandbox.GetJobs()...)

	// operations
	providers = append(providers, operations.GetJobs()...)
	providers = append(providers, fx.Provide(jobs.AsJobHandler("operations", check.New)))

	// sync
	providers = append(providers, sync.GetJobs()...)

	// actions
	providers = append(providers, actions.GetJobs()...)

	// org-only providers
	providers = append(providers, build.GetJobs()...)

	// install-only proviers
	providers = append(providers, deploy.GetJobs()...)

	providers = append(
		providers,
		[]fx.Option{
			// provide process for the heartbeater
			// NOTE(fd): this process uses the empty string
			fx.Supply(fx.Annotate("", fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			fx.Invoke(jobloop.WithOperationsJobLoops(func([]jobloop.JobLoop) {})),

			// registry and heartbeater
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
			fx.Invoke(func(*registry.Registry) {}),
		}...,
	)

	// NOTE(fd): we need a way to determine what kind of runner we are running as
	fx.New(providers...).Run()
}
