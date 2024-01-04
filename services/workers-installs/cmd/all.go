package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/sender"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/deprovision"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/dns"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/provision"
	"github.com/powertoolsdev/mono/services/workers-installs/internal/runner"
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
		n, err = sender.NewSlackSender(cfg.InstallationBotsSlackWebhookURL)
		if err != nil {
			n = sender.NewNoopSender()
		}
	}

	prWorkflow := provision.NewWorkflow(cfg)
	prRWorkflow := runner.NewWorkflow(cfg)
	dprWorkflow := deprovision.NewWorkflow(cfg)
	dnsWorkflow := dns.NewWorkflow(cfg)

	v := validator.New()
	wkr, err := worker.New(v, worker.WithConfig(&cfg.Config),
		// register workflows
		worker.WithWorkflow(prWorkflow.Provision),
		worker.WithWorkflow(dprWorkflow.Deprovision),
		worker.WithWorkflow(prRWorkflow.ProvisionRunner),
		worker.WithWorkflow(dnsWorkflow.ProvisionDNS),

		// register activities
		worker.WithActivity(provision.NewActivities(v, cfg, n)),
		worker.WithActivity(runner.NewActivities(v, cfg)),
		worker.WithActivity(deprovision.NewActivities(v, n, &cfg)),
		worker.WithActivity(dns.NewActivities(v)),
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
