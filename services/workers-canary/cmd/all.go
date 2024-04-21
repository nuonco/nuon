package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/services/config"
	tmetrics "github.com/powertoolsdev/mono/pkg/temporal/metrics"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/workflows"
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

	v := validator.New()
	tmetricsWriter, err := tmetrics.New(v, tmetrics.WithTags(map[string]string{
		"git_ref":   cfg.GitRef,
		"service":   cfg.ServiceName,
		"namespace": cfg.TemporalNamespace,
	}))
	if err != nil {
		log.Fatalf("unable to create new temporal metrics writer: %s", err.Error())
	}

	wkflow, err := workflows.New(v, cfg, tmetricsWriter)
	if err != nil {
		log.Fatalf("unable to create workflows: %s", err.Error())
	}

	acts, err := activities.New(v,
		activities.WithConfig(&cfg),
	)
	if err != nil {
		log.Fatalf("unable to create activities: %s", err.Error())
	}

	sharedActs, err := sharedactivities.New(v,
		sharedactivities.WithTemporalHost(cfg.TemporalHost),
	)
	if err != nil {
		log.Fatalf("unable to create activities: %s", err.Error())
	}

	wkr, err := worker.New(v, worker.WithConfig(&cfg.Config),
		worker.WithWorkflow(wkflow.Provision),
		worker.WithWorkflow(wkflow.Deprovision),

		// register activities
		worker.WithActivity(acts),
		worker.WithActivity(sharedActs),
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
