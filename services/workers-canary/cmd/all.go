package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/spf13/cobra"
	tworker "go.temporal.io/sdk/worker"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all workers",
	Run:   runAll,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(allCmd)
}

func runAll(cmd *cobra.Command, _ []string) {
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("unable to validate config: %v", err)
	}

	wkr, err := worker.New(validator.New(), worker.WithConfig(&cfg.Config)) //// register workflows
	//worker.WithWorkflow(stWkflow.Start),
	//worker.WithWorkflow(bldWkflow.Build),
	//worker.WithWorkflow(instWkflow.ProvisionInstances),

	// register activities
	//worker.WithActivity(start.NewActivities()),
	//worker.WithActivity(instances.NewActivities(cfg)),

	if err != nil {
		log.Fatalf("unable to initialize worker: %s", err.Error())
	}

	interruptCh := tworker.InterruptCh()
	err = wkr.Run(interruptCh)
	if err != nil {
		log.Fatalf("unable to run worker: %v", err)
	}
}
