package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/workflows/worker"
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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
)

var namespace string

func (c *cli) registerWorker() error {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "run worker",
		Run:   c.runWorker,
	}
	rootCmd.AddCommand(cmd)
	helpText := "namespace defines the namespace whose workers to run. e.g. all, general, orgs, apps, components, installs, releases."
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "all", helpText)
	return nil
}

func (c *cli) runWorker(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		fx.Invoke(db.DBGroupParam(func(dbs []*gorm.DB) {})),
	}

	// generals worker
	if namespace == "all" || namespace == "general" {
		providers = append(providers,
			fx.Provide(generalactivities.New),
			fx.Provide(generalworker.NewWorkflows),
			fx.Provide(generalworker.New),
			fx.Invoke(func(*generalworker.Worker) {}),
		)
	}

	// orgs worker
	if namespace == "all" || namespace == "orgs" {
		providers = append(providers,
			fx.Provide(orgsactivities.New),
			fx.Provide(orgsworker.NewWorkflows),
			fx.Provide(orgsworker.New),
			fx.Invoke(func(*orgsworker.Worker) {}),
		)
	}

	// apps worker
	if namespace == "all" || namespace == "apps" {
		providers = append(providers,
			fx.Provide(appsactivities.New),
			fx.Provide(appsworker.NewWorkflows),
			fx.Provide(appsworker.New),
			fx.Invoke(func(*appsworker.Worker) {}),
		)
	}

	// components worker
	if namespace == "all" || namespace == "components" {
		providers = append(providers,
			fx.Provide(componentsactivities.New),
			fx.Provide(componentsworker.NewWorkflows),
			fx.Provide(componentsworker.New),
			fx.Invoke(func(*componentsworker.Worker) {}),
		)
	}

	// installs worker
	if namespace == "all" || namespace == "installs" {
		providers = append(providers,
			fx.Provide(installsactivities.New),
			fx.Provide(installsworker.NewWorkflows),
			fx.Provide(installsworker.New),
			fx.Invoke(func(*installsworker.Worker) {}),
		)
	}

	if namespace == "all" || namespace == "releases" {
		providers = append(providers,
			fx.Provide(releasesactivities.New),
			fx.Provide(releasesworker.NewWorkflows),
			fx.Provide(releasesworker.New),
			fx.Invoke(func(*releasesworker.Worker) {}),
		)
	}

	if namespace == "all" || namespace == "runners" {
		providers = append(providers,
			// runners worker
			fx.Provide(runnersactivities.New),
			fx.Provide(runnersworker.NewWorkflows),
			fx.Provide(worker.AsWorker(runnersworker.New)),
		)
	}

	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
