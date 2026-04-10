package cmd

import (
	"slices"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/actions"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/build"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/deploy"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/operations"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandbox"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sync"

	"github.com/nuonco/nuon/bins/runner/internal/registry"
	"github.com/nuonco/nuon/bins/runner/internal/sandboxctl"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/jobloop"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"

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
			// derive process type from settings groups:
			// org runners have "build" in groups, install runners have "deploys"
			fx.Provide(fx.Annotate(func(s *settings.Settings) string {
				if slices.Contains(s.Groups, "deploys") {
					return "install"
				}
				return "build"
			}, fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			fx.Invoke(jobloop.WithOperationsJobLoops(func([]jobloop.JobLoop) {})),

			// sandbox control API
			fx.Provide(sandboxctl.New),
			fx.Invoke(func(*sandboxctl.Server) {}),

			// registry, heartbeater, process registrar, and shutdown poller
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
			fx.Invoke(func(*process.Registrar) {}),
			fx.Invoke(func(*process.ShutdownPoller) {}),
			fx.Invoke(func(*registry.Registry) {}),
		}...,
	)

	// NOTE(fd): we need a way to determine what kind of runner we are running as
	fx.New(providers...).Run()
}
