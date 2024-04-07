package cmd

import (
	appsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	appsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/activities"
	componentsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
	componentsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	installsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
	installsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	orgsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	releasesworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	releasesactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker/activities"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func (c *cli) registerWorker() error {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "run worker",
		Run:   c.runWorker,
	}
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runWorker(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		fx.Invoke(func(*gorm.DB) {}),

		// orgs worker
		fx.Provide(orgsactivities.New),
		fx.Provide(orgsworker.NewWorkflows),
		fx.Provide(orgsworker.New),
		fx.Invoke(func(*orgsworker.Worker) {
		}),

		// apps worker
		fx.Provide(appsactivities.New),
		fx.Provide(appsworker.NewWorkflows),
		fx.Provide(appsworker.New),
		fx.Invoke(func(*appsworker.Worker) {
		}),

		// components worker
		fx.Provide(componentsactivities.New),
		fx.Provide(componentsworker.NewWorkflows),
		fx.Provide(componentsworker.New),
		fx.Invoke(func(*componentsworker.Worker) {
		}),

		// installs worker
		fx.Provide(installsactivities.New),
		fx.Provide(installsworker.NewWorkflows),
		fx.Provide(installsworker.New),
		fx.Invoke(func(*installsworker.Worker) {
		}),

		// releases worker
		fx.Provide(releasesactivities.New),
		fx.Provide(releasesworker.NewWorkflows),
		fx.Provide(releasesworker.New),
		fx.Invoke(func(*releasesworker.Worker) {
		}),
	}
	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
