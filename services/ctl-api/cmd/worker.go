package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/profiles"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	actionsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker"
	actionsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	appbranchesworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/worker"
	appbranchesactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/worker/activities"
	appsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	appsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	componentsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
	componentsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	generalworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker"
	generalactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	installsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
	installsactionsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/actions"
	installsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	installscomponentsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/components"
	installssandboxworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/sandbox"
	installsstackworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/stack"
	installstate "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/state"
	orgsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	releasesworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	releasesactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	runnersworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker"
	runnersactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/interceptors"
	metricsinterceptor "github.com/powertoolsdev/mono/services/ctl-api/internal/interceptors/metrics"
	validateinterceptor "github.com/powertoolsdev/mono/services/ctl-api/internal/interceptors/validate"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
	flowactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
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
	providers := []fx.Option{
		fx.Provide(interceptors.AsInterceptor(metricsinterceptor.New)),
		fx.Provide(interceptors.AsInterceptor(validateinterceptor.New)),
	}
	providers = append(providers, c.providers()...)

	profilerOptions := profiles.LoadOptionsFromEnv()
	providers = append(providers, profiles.Module(profilerOptions))

	// shared activities and workflows
	providers = append(
		providers,
		fx.Provide(jobactivities.New),
		fx.Provide(flowactivities.New),
		fx.Provide(signalsactivities.New),
		fx.Provide(statusactivities.New),
		fx.Provide(activities.New),
		fx.Provide(job.New),
		fx.Provide(workflows.NewActivities),
		fx.Provide(workflows.NewWorkflows),
	)

	// generals worker
	if (namespace == "all" || namespace == "general") && !shouldSkipNamespace("general") {
		providers = append(providers,
			fx.Provide(generalactivities.New),
			fx.Provide(generalworker.NewWorkflows),
			fx.Provide(worker.AsWorker(generalworker.New)),
		)
	}

	// orgs worker
	if (namespace == "all" || namespace == "orgs") && !shouldSkipNamespace("orgs") {
		providers = append(providers,
			fx.Provide(orgsactivities.New),
			fx.Provide(orgsworker.NewWorkflows),
			fx.Provide(worker.AsWorker(orgsworker.New)),
		)
	}

	// apps worker
	if (namespace == "all" || namespace == "apps") && !shouldSkipNamespace("apps") {
		providers = append(providers,
			fx.Provide(appsactivities.New),
			fx.Provide(appsworker.NewWorkflows),
			fx.Provide(worker.AsWorker(appsworker.New)))
	}

	// app-branches worker
	if (namespace == "all" || namespace == "app-branches") && !shouldSkipNamespace("app-branches") {
		providers = append(providers,
			fx.Provide(appbranchesactivities.New),
			fx.Provide(appbranchesworker.NewWorkflows),
			fx.Provide(worker.AsWorker(componentsworker.New)),
		)
	}

	// components worker
	if (namespace == "all" || namespace == "components") && !shouldSkipNamespace("components") {
		providers = append(providers,
			fx.Provide(componentsactivities.New),
			fx.Provide(componentsworker.NewWorkflows),
			fx.Provide(worker.AsWorker(componentsworker.New)),
		)
	}

	// installs worker
	if (namespace == "all" || namespace == "installs") && !shouldSkipNamespace("installs") {
		providers = append(providers,
			fx.Provide(installsactivities.New),
			fx.Provide(installsworker.NewWorkflows),
			fx.Provide(installsactionsworker.NewWorkflows),
			fx.Provide(installscomponentsworker.NewWorkflows),
			fx.Provide(installssandboxworker.NewWorkflows),
			fx.Provide(installsstackworker.NewWorkflows),
			fx.Provide(installstate.New),
			fx.Provide(worker.AsWorker(installsworker.New)),
		)
	}

	if (namespace == "all" || namespace == "releases") && !shouldSkipNamespace("releases") {
		providers = append(providers,
			fx.Provide(releasesactivities.New),
			fx.Provide(releasesworker.NewWorkflows),
			fx.Provide(worker.AsWorker(releasesworker.New)),
		)
	}

	if (namespace == "all" || namespace == "runners") && !shouldSkipNamespace("runners") {
		providers = append(providers,
			// runners worker
			fx.Provide(runnersactivities.New),
			fx.Provide(runnersworker.NewWorkflows),
			fx.Provide(worker.AsWorker(runnersworker.New)),
		)
	}

	if (namespace == "all" || namespace == "actions") && !shouldSkipNamespace("actions") {
		providers = append(providers,
			// actions worker
			fx.Provide(actionsactivities.New),
			fx.Provide(actionsworker.NewWorkflows),
			fx.Provide(worker.AsWorker(actionsworker.New)),
		)
	}

	providers = append(providers,
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {
		})),
	)
	fx.New(providers...).Run()
}
