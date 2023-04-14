package cmd

import (
	"log"

	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createdeployment"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
	"github.com/spf13/cobra"
	tworker "go.temporal.io/sdk/worker"
)

var runWorkerCmd = &cobra.Command{
	Use:   "workers",
	Short: "run background workers",
	Run:   runWorkers,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runWorkerCmd)
}

func runWorkers(cmd *cobra.Command, _ []string) {
	app, err := newApp(cmd.Flags())
	if err != nil {
		log.Fatalf("unable to load worker: %s", err)
	}

	createAppJob := createapp.New(app.v)
	createOrgJob := createorg.New(app.v)
	createInstallJob := createinstall.New(app.v)
	createDeploymentJob := createdeployment.New(app.v)

	sharedActs, err := sharedactivities.New(app.v)
	if err != nil {
		log.Fatalf("unable to load shared activities: %s", err)
	}

	wkr, err := worker.New(app.v, worker.WithConfig(&app.cfg.Config),
		// register workflows
		worker.WithWorkflow(createAppJob.CreateApp),
		worker.WithWorkflow(createOrgJob.CreateOrg),
		worker.WithWorkflow(createInstallJob.CreateInstall),
		worker.WithWorkflow(createDeploymentJob.CreateDeployment),

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
