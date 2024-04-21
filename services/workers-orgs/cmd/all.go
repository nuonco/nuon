package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/pkg/sender"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/deprovision"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/iam"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/kms"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/provision"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/runner"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/server"
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
	case config.Development:
		n = sender.NewNoopSender()
	default:
		n, err = sender.NewSlackSender(cfg.BotsSlackWebhookURL)
		if err != nil {
			n = sender.NewNoopSender()
		}
	}

	wkflow := provision.NewWorkflow(cfg)
	tdWkflow := deprovision.NewWorkflow(cfg)
	runiFlow := runner.NewWorkflow(cfg)
	srvWkflow := server.NewWorkflow(cfg)
	iamWkflow := iam.NewWorkflow(cfg)

	kmsWkflow := kms.NewWorkflow(cfg)

	v := validator.New()
	wkr, err := worker.New(v, worker.WithConfig(&cfg.Config),
		// register workflows
		worker.WithWorkflow(wkflow.Provision),
		worker.WithWorkflow(tdWkflow.Deprovision),
		worker.WithWorkflow(iamWkflow.DeprovisionIAM),
		worker.WithWorkflow(runiFlow.ProvisionRunner),
		worker.WithWorkflow(srvWkflow.ProvisionServer),
		worker.WithWorkflow(iamWkflow.ProvisionIAM),
		worker.WithWorkflow(kmsWkflow.ProvisionKMS),

		// register activities
		worker.WithActivity(provision.NewActivities(n)),
		worker.WithActivity(runner.NewActivities(v, cfg)),
		worker.WithActivity(deprovision.NewActivities()),
		worker.WithActivity(server.NewActivities(v)),
		worker.WithActivity(iam.NewActivities()),
		worker.WithActivity(kms.NewActivities()),
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
