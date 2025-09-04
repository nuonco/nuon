package cmd

import (
	"github.com/powertoolsdev/mono/bins/runner/internal/jobs/management"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/heartbeater"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/jobloop"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerMng() error {
	mngCmd := &cobra.Command{
		Use:     "mng",
		Short:   "Run in management mode.",
		Long:    "Run in management mode and oversee an install mode process in a standalone VM.",
		Aliases: []string{"management"},
		Run:     c.runMng,
	}

	rootCmd.AddCommand(mngCmd)
	return nil
}

func (c *cli) runMng(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{fx.Provide(log.NewSystem)}
	providers = append(c.commonProviders(), providers...)
	providers = append(providers, management.GetJobs()...)
	// add mng and heartbeater to the mng process
	providers = append(providers,
		[]fx.Option{
			// provide process for the heartbeater
			fx.Supply(fx.Annotate("mng", fx.ResultTags(`name:"process"`))),
			// start all job loops
			fx.Invoke(jobloop.WithJobLoops(func([]jobloop.JobLoop) {})),
			// NOTE: we do not include the `operations` job loops here
			// start registry and heartbeater
			fx.Invoke(func(*heartbeater.HeartBeater) {}),
		}...,
	)
	// run
	fx.New(providers...).Run()
}
