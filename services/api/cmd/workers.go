package cmd

import (
	"log"

	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	temporalclient "github.com/powertoolsdev/mono/services/api/internal/clients/temporal"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build"
	buildactivities "github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createdeployment"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteorg"
	"github.com/spf13/cobra"
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
	deleteOrgJob := deleteorg.New(app.v)
	buildJob := build.New(build.Config{Config: app.cfg.Config})
	orgsTc, err := temporalclient.New(temporalclient.WithConfig(app.cfg), temporalclient.WithNamespace("orgs"))
	if err != nil {
		log.Fatalf("unable to create orgs temporal client for background activities: %s", err)
	}
	appsTc, err := temporalclient.New(temporalclient.WithConfig(app.cfg), temporalclient.WithNamespace("apps"))
	if err != nil {
		log.Fatalf("unable to create apps temporal client for background activities: %s", err)
	}
	installsTc, err := temporalclient.New(temporalclient.WithConfig(app.cfg), temporalclient.WithNamespace("installs"))
	if err != nil {
		log.Fatalf("unable to create installs temporal client for background activities: %s", err)
	}
	deploymentsTc, err := temporalclient.New(temporalclient.WithConfig(app.cfg), temporalclient.WithNamespace("deployments"))
	if err != nil {
		log.Fatalf("unable to create deployments temporal client for background activities: %s", err)
	}
	createOrgActivites := createorg.NewActivities(orgsTc)
	deleteOrgActivites := deleteorg.NewActivities(orgsTc)
	createAppActivites := createapp.NewActivities(app.db, appsTc)
	createInstallActivites := createinstall.NewActivities(app.db, installsTc)
	deleteInstallActivites := deleteinstall.NewActivities(app.db, installsTc)
	createDeploymentActivites := createdeployment.NewActivities(app.db, deploymentsTc)
	buildActivities := buildactivities.New(app.db)

	createInstallJob := createinstall.New(app.v)
	deleteInstallJob := deleteinstall.New(app.v)
	createDeploymentJob := createdeployment.New(app.v)

	sharedActs, err := sharedactivities.New(app.v, sharedactivities.WithTemporalHost(app.cfg.TemporalHost))
	if err != nil {
		log.Fatalf("unable to load shared activities: %s", err)
	}

	wkr, err := worker.New(app.v, worker.WithConfig(&app.cfg.Config),
		// register workflows
		worker.WithWorkflow(createAppJob.CreateApp),
		worker.WithWorkflow(createOrgJob.CreateOrg),
		worker.WithWorkflow(deleteOrgJob.DeleteOrg),
		worker.WithWorkflow(createInstallJob.CreateInstall),
		worker.WithWorkflow(deleteInstallJob.DeleteInstall),
		worker.WithWorkflow(createDeploymentJob.CreateDeployment),
		worker.WithWorkflow(buildJob.Build),

		// register activites
		worker.WithActivity(createOrgActivites),
		worker.WithActivity(deleteOrgActivites),
		worker.WithActivity(createAppActivites),
		worker.WithActivity(createInstallActivites),
		worker.WithActivity(deleteInstallActivites),
		worker.WithActivity(createDeploymentActivites),
		worker.WithActivity(sharedActs),
		worker.WithActivity(buildActivities),
	)
	if err != nil {
		log.Fatalf("unable to initialize worker: %s", err.Error())
	}
	interruptCh := make(chan interface{})
	wkr.Run(interruptCh)
	if err != nil {
		log.Fatalf("unable to run worker: %v", err)
	}
}
