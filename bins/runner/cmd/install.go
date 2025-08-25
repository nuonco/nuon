package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/actions"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/deploy"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/operations"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sandbox"
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/sync"

	"github.com/powertoolsdev/mono/bins/runner/internal/registry"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"

	check "github.com/powertoolsdev/mono/bins/runner/internal/jobs/healthcheck/check"
)

func (c *cli) registerInstall() error {
	runCmd := &cobra.Command{
		Use:   "install",
		Short: "Run in install mode.",
		Long:  "Run in install mode and handle sandbox, component, and action jobs.",
		Run:   c.runInstall,
	}

	rootCmd.AddCommand(runCmd)
	return nil
}

func (c *cli) runInstall(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{}

	// common providers
	providers = append(providers, c.providers()...)

	// operations
	providers = append(providers, operations.GetJobs()...)
	providers = append(providers, fx.Provide(jobs.AsJobHandler("operations", check.New)))

	// install-mode providers
	providers = append(providers, sync.GetJobs()...)
	providers = append(providers, sandbox.GetJobs()...)
	providers = append(providers, deploy.GetJobs()...)
	providers = append(providers, actions.GetJobs()...)

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
