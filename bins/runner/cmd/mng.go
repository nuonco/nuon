package cmd

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/management"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerMng() error {
	mngCmd := &cobra.Command{
		Use:  "mng",
		Long: "mng executes the runner management mode to start and monitor a runner.",
		Run:  c.runMng,
	}

	rootCmd.AddCommand(mngCmd)
	return nil
}

func (c *cli) runMng(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{fx.Provide(log.NewSystem)}

	providers = append(c.commonProviders(), providers...)

	providers = append(providers, management.GetJobs()...)

	fx.New(providers...).Run()
}
