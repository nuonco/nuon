package cmd

import (
	orgsworker "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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
		// orgs workflows
		fx.Provide(orgsactivities.New),
		fx.Provide(orgsworker.NewWorkflows),
		fx.Provide(orgsworker.New),
		fx.Invoke(func(*orgsworker.Worker) {
		}),
	}
	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
