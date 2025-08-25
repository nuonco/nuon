package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/build"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync"

	"github.com/powertoolsdev/mono/bins/runner/internal/registry"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"

	check "github.com/powertoolsdev/mono/bins/runner/internal/jobs/healthcheck/check"
)

func (c *cli) registerBuild() error {
	runCmd := &cobra.Command{
		Use:     "build",
		Short:   "Run in org/build mode.",
		Long:    "Run in org mode and handle component builds.",
		Aliases: []string{"org"},
		Run:     c.runBuild,
	}

	rootCmd.AddCommand(runCmd)
	return nil
}

func (c *cli) runBuild(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{}

	// common providers
	providers = append(providers, c.providers()...)

	// operations
	providers = append(providers, operations.GetJobs()...)
	providers = append(providers, fx.Provide(jobs.AsJobHandler("operations", check.New)))

	// org-mode providers
	providers = append(providers, sync.GetJobs()...)
	providers = append(providers, build.GetJobs()...)

	// heartbeat, registry, job loop execution
	providers = append(
		providers,
		[]fx.Option{
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			fx.Invoke(jobloop.WithOperationsJobLoops(func([]jobloop.JobLoop) {})),

			// registry and heartbeater
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
			fx.Invoke(func(*registry.Registry) {}),
		}...,
	)

	// run
	fx.New(providers...).Run()
}
