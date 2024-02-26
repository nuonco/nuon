package cmd

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	shared "github.com/powertoolsdev/mono/services/workers-apps/internal"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/deprovision"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/repository"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/sync"
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

	// parent workflows
	prWkflow := provision.NewWorkflow(cfg)
	dprWkflow := deprovision.NewWorkflow(v, cfg)
	syncWkflow := sync.NewWorkflow(v, cfg)

	// child workflows
	pwkflow := project.NewWorkflow(cfg)
	rwkflow := repository.NewWorkflow(cfg)

	wkr, err := worker.New(v, worker.WithConfig(&cfg.Config),
		// register workflows
		worker.WithWorkflow(prWkflow.Provision),
		worker.WithWorkflow(dprWkflow.Deprovision),
		worker.WithWorkflow(syncWkflow.Sync),
		worker.WithWorkflow(pwkflow.ProvisionProject),
		worker.WithWorkflow(rwkflow.ProvisionRepository),

		// register activities
		worker.WithActivity(provision.NewActivities()),
		worker.WithActivity(deprovision.NewActivities(v, cfg)),
		worker.WithActivity(project.NewActivities(v)),
		worker.WithActivity(repository.NewActivities()),
		worker.WithActivity(sync.NewActivities(v, cfg)),
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
