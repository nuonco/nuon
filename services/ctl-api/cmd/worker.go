package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/profiles"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
	"github.com/nuonco/nuon/services/ctl-api/internal/health"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
)

var (
	namespace      string
	skipNamespaces string
)

func (c *cli) registerWorker() error {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "run worker",
		Run:   c.runWorker,
	}
	rootCmd.AddCommand(cmd)
	helpText := "namespace defines the namespace whose workers to run. e.g. all, general, orgs, apps, components, installs, releases."
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "all", helpText)
	rootCmd.PersistentFlags().StringVar(&skipNamespaces, "skip", "", "comma-separated list of namespaces to skip (e.g. 'installs,releases')")
	return nil
}

// shouldSkipNamespace checks if a namespace should be skipped based on the skipNamespaces flag
func shouldSkipNamespace(ns string) bool {
	if skipNamespaces == "" {
		return false
	}

	skips := strings.Split(skipNamespaces, ",")
	for _, skip := range skips {
		if strings.TrimSpace(skip) == ns {
			return true
		}
	}
	return false
}

func (c *cli) runWorker(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{}
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// Add worker interceptors and shared workflows. SlackLibsModule supplies
	// *slackclient.Client to the Slack signal lifecycle hook in
	// SharedWorkflowsModule; without it the hook's Supports() short-circuits
	// because SlackParams.SlackClient is optional and resolves to nil.
	providers = append(providers,
		fxmodules.WorkerInterceptorsModule,
		fxmodules.SharedWorkflowsModule,
		fxmodules.SlackLibsModule,
	)

	// Add namespace-specific worker modules based on flags
	if (namespace == "all" || namespace == "general") && !shouldSkipNamespace("general") {
		providers = append(providers, fxmodules.GeneralWorkerModule)
	}

	if (namespace == "all" || namespace == "orgs") && !shouldSkipNamespace("orgs") {
		providers = append(providers, fxmodules.OrgsWorkerModule)
	}

	if (namespace == "all" || namespace == "apps") && !shouldSkipNamespace("apps") {
		providers = append(providers, fxmodules.AppsWorkerModule)
	}

	if (namespace == "all" || namespace == "components") && !shouldSkipNamespace("components") {
		providers = append(providers, fxmodules.ComponentsWorkerModule)
	}

	runInstalls := (namespace == "all" || namespace == "installs") && !shouldSkipNamespace("installs")
	runInstallCrons := (namespace == "all" || namespace == "install-crons") && !shouldSkipNamespace("install-crons")
	if runInstalls || runInstallCrons {
		providers = append(providers, fxmodules.InstallWorkerProvidersModule)
	}
	if runInstalls {
		providers = append(providers, fxmodules.InstallsWorkerModule)
	}
	if runInstallCrons {
		providers = append(providers, fxmodules.InstallCronWorkerModule)
	}

	// Releases namespace removed - being deprecated
	// if (namespace == "all" || namespace == "releases") && !shouldSkipNamespace("releases") {
	// 	providers = append(providers, fxmodules.ReleasesWorkerModule)
	// }

	runRunners := (namespace == "all" || namespace == "runners") && !shouldSkipNamespace("runners")
	runRunnerHealthcheckCrons := (namespace == "all" || namespace == "runner-healthcheck-crons") && !shouldSkipNamespace("runner-healthcheck-crons")
	if runRunners || runRunnerHealthcheckCrons {
		providers = append(providers, fxmodules.RunnerWorkerProvidersModule)
	}
	if runRunners {
		providers = append(providers, fxmodules.RunnersWorkerModule)
	}
	if runRunnerHealthcheckCrons {
		providers = append(providers, fxmodules.RunnerHealthcheckCronWorkerModule)
	}

	if (namespace == "all" || namespace == "actions") && !shouldSkipNamespace("actions") {
		providers = append(providers, fxmodules.ActionsWorkerModule)
	}

	if (namespace == "all" || namespace == "onboardings") && !shouldSkipNamespace("onboardings") {
		providers = append(providers, fxmodules.OnboardingsWorkerModule)
	}

	if (namespace == "all" || namespace == "vcs") && !shouldSkipNamespace("vcs") {
		providers = append(providers, fxmodules.VCSWorkerModule)
	}

	// Add worker healthcheck server
	providers = append(providers,
		fx.Provide(health.NewWorkerHealthcheck),
		fx.Invoke(func(lc fx.Lifecycle, hc *health.WorkerHealthcheckServer) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					return hc.Start()
				},
				OnStop: func(ctx context.Context) error {
					return hc.Stop(ctx)
				},
			})
		}),
	)

	// Add final invocations
	providers = append(providers,
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {})),
	)

	fx.New(providers...).Run()
}
