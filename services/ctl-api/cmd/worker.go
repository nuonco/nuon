package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/profiles"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	actionsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker"
	actionsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	appsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	appsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	componentsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
	componentsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	generalworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker"
	generalactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	installsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
	installsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
)

var namespace, mode string

func (c *cli) registerWorker() error {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "run worker",
		Run:   c.runWorker,
	}
	rootCmd.AddCommand(cmd)
	helpText := "namespace defines the namespace whose workers to run. e.g. all, general, orgs, apps, components, installs, releases."
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "all", helpText)
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "all", "mode of the worker. Options: all, activities, workflows")

	return nil
}

func provideWorkerMode() worker.Mode {
	if mode != "all" && mode != "activities" && mode != "workflows" {
		mode = "all"
	}
	return worker.Mode(mode)
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
		fx.Provide(signalsactivities.New),
		fx.Provide(activities.New),
		fx.Provide(job.New),
		fx.Provide(workflows.NewActivities),
		fx.Provide(workflows.NewWorkflows),
	)

	// generals worker
	if namespace == "all" || namespace == "general" {
		providers = append(providers,
			fx.Provide(generalactivities.New),
			fx.Provide(generalworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(generalworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(generalworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(generalworker.NewActivityWorker)),
			)
		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(generalworker.NewWorkflowWorker)),
			)
		}

	}

	// orgs worker
	if namespace == "all" || namespace == "orgs" {
		providers = append(providers,
			fx.Provide(orgsactivities.New),
			fx.Provide(orgsworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(orgsworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(orgsworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(orgsworker.NewActivityWorker)),
			)

		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(orgsworker.NewWorkflowWorker)),
			)
		}

	}

	// apps worker
	if namespace == "all" || namespace == "apps" {
		providers = append(providers,
			fx.Provide(appsactivities.New),
			fx.Provide(appsworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(appsworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(appsworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(appsworker.NewActivityWorker)),
			)
		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(appsworker.NewWorkflowWorker)),
			)
		}

	}

	if namespace == "all" || namespace == "components" {
		providers = append(providers,
			fx.Provide(componentsactivities.New),
			fx.Provide(componentsworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(componentsworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(componentsworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(componentsworker.NewActivityWorker)),
			)
		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(componentsworker.NewWorkflowWorker)),
			)
		}

	}

	// installs worker
	if namespace == "all" || namespace == "installs" {
		providers = append(providers,
			fx.Provide(installsactivities.New),
			fx.Provide(installsworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(installsworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(installsworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(installsworker.NewActivityWorker)),
			)

		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(installsworker.NewWorkflowWorker)),
			)
		}

	}

	if namespace == "all" || namespace == "releases" {
		providers = append(providers,
			fx.Provide(releasesactivities.New),
			fx.Provide(releasesworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(releasesworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(releasesworker.NewWorkflowWorker)),
			)

		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(releasesworker.NewActivityWorker)),
			)

		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(releasesworker.NewWorkflowWorker)),
			)
		}

	}

	if namespace == "all" || namespace == "runners" {
		providers = append(providers,
			fx.Provide(runnersactivities.New),
			fx.Provide(runnersworker.NewWorkflows),
		)
		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(runnersworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(runnersworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(runnersworker.NewActivityWorker)),
			)
		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(runnersworker.NewWorkflowWorker)),
			)
		}

	}

	if namespace == "all" || namespace == "actions" {
		providers = append(providers,
			fx.Provide(actionsactivities.New),
			fx.Provide(actionsworker.NewWorkflows),
		)

		switch provideWorkerMode() {
		case worker.ModeAll, worker.Mode(""):
			providers = append(providers,
				fx.Provide(worker.AsWorker(actionsworker.NewActivityWorker)),
				fx.Provide(worker.AsWorker(actionsworker.NewWorkflowWorker)),
			)
		case worker.ModeActivities:
			providers = append(providers,
				fx.Provide(worker.AsWorker(actionsworker.NewActivityWorker)),
			)
		case worker.ModeWorkflows:
			providers = append(providers,
				fx.Provide(worker.AsWorker(actionsworker.NewWorkflowWorker)),
			)
		}

	}

	providers = append(providers,
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {
		})),
	)
	fx.New(providers...).Run()
}
