package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/sender"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-instances/internal"
	"github.com/powertoolsdev/mono/services/workers-instances/internal/provision"
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

	var (
		n   sender.NotificationSender
		err error
	)
	switch cfg.Env {
	case config.Local, config.Development:
		n = sender.NewNoopSender()
	default:
		n, err = sender.NewSlackSender(cfg.DeploymentBotsSlackWebhookURL)
		if err != nil {
			n = sender.NewNoopSender()
		}
	}

	wkflow := provision.NewWorkflow(cfg)

	v := validator.New()
	wkr, err := worker.New(v, worker.WithConfig(&cfg.Config),
		// register workflows
		worker.WithWorkflow(wkflow.Provision),

		// register activities
		worker.WithActivity(provision.NewActivities(v, n)),
	)
	if err != nil {
		log.Fatalf("unable to initialize worker: %s", err.Error())
	}

	interruptCh := tworker.InterruptCh()
	err = wkr.Run(interruptCh)
	if err != nil {
		log.Fatalf("unable to run worker: %v", err)
	}
}
