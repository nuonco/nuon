package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/actions"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/deploy"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/operations"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sync"
	"github.com/nuonco/nuon/pkg/runner/jobs"

	"github.com/nuonco/nuon/bins/runner/internal/registry"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/componenthealth"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/telemetryexport"

	check "github.com/nuonco/nuon/bins/runner/internal/jobs/healthcheck/check"
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

func (c *cli) runRun(_ *cobra.Command, _ []string) {
	fx.New(c.runOptions()...).Run()
}

func (c *cli) runOptions() []fx.Option {
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

	// deploy providers
	providers = append(providers, deploy.GetJobs()...)
	providers = append(providers, audit.Module, telemetryexport.Module)

	providers = append(
		providers,
		[]fx.Option{
			fx.Supply(fx.Annotate("install", fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			fx.Invoke(jobloop.WithOperationsJobLoops(func([]jobloop.JobLoop) {})),

			// sandbox control API

			// registry, heartbeater, process registrar, and shutdown poller
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
			fx.Invoke(func(*process.Registrar) {}),
			fx.Invoke(func(*process.ShutdownPoller) {}),
			fx.Invoke(func(*registry.Registry) {}),

			// component health watch engine (no-op unless install process)
			fx.Provide(componenthealth.New),
			fx.Invoke(func(*componenthealth.Engine) {}),
		}...,
	)

	return providers
}
