package cmd

import (
	"fmt"

	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	metaactivities "github.com/powertoolsdev/mono/pkg/workflows/meta"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build"
	buildactivities "github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteorg"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/startdeploy"
	startdeployacts "github.com/powertoolsdev/mono/services/api/internal/jobs/startdeploy/activities"
)

func (a *app) buildsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	wkflow := build.New(a.cfg)
	acts := buildactivities.New(a.db, a.cfg.GithubAppID, a.cfg.GithubAppKeySecretName)
	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithNamespace("builds"),
		worker.WithWorkflow(wkflow.Build),
		worker.WithActivity(acts),
		worker.WithActivity(sa),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create builds worker: %w", err)
	}

	return wkr, nil
}

func (a *app) deploysWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	cfg := startdeploy.Config{
		OrgsDeploymentsRoleTemplate: a.cfg.OrgsDeploymentsRoleTemplate,
		DeploymentsBucket:           a.cfg.DeploymentsBucket,
	}
	startDeployWkflow := startdeploy.New(a.v, cfg)
	startdeployActs := startdeployacts.NewActivities(a.v, a.db)
	metaStartActivity := metaactivities.NewStartActivity()
	metaEndActivity := metaactivities.NewFinishActivity()
	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithWorkflow(startDeployWkflow.StartDeploy),
		worker.WithNamespace("deploys"),
		worker.WithActivity(startdeployActs),
		worker.WithActivity(sa),
		worker.WithActivity(metaStartActivity),
		worker.WithActivity(metaEndActivity),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create apps worker: %w", err)
	}

	return wkr, nil
}

func (a *app) appsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	createWkflow := createapp.New(a.v)
	createActs := createapp.NewActivities(a.db, a.wfc)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithWorkflow(createWkflow.CreateApp),
		worker.WithNamespace("apps"),
		worker.WithActivity(createActs),
		worker.WithActivity(sa),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create apps worker: %w", err)
	}

	return wkr, nil
}

func (a *app) orgsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	createWkflow := createorg.New(a.v)
	deleteWkflow := deleteorg.New(a.v)
	createActs := createorg.NewActivities(a.wfc)
	deleteActs := deleteorg.NewActivities(a.wfc)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithNamespace("orgs"),

		worker.WithWorkflow(createWkflow.CreateOrg),
		worker.WithWorkflow(deleteWkflow.DeleteOrg),

		worker.WithActivity(createActs),
		worker.WithActivity(deleteActs),
		worker.WithActivity(sa),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create orgs worker: %w", err)
	}

	return wkr, nil
}

func (a *app) installsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	createWkflow := createinstall.New(a.v)
	deleteWkflow := deleteinstall.New(a.v)
	createActs := createinstall.NewActivities(a.db, a.wfc)
	deleteActs := deleteinstall.NewActivities(a.db, a.wfc)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithNamespace("installs"),

		worker.WithWorkflow(createWkflow.CreateInstall),
		worker.WithWorkflow(deleteWkflow.DeleteInstall),

		worker.WithActivity(createActs),
		worker.WithActivity(deleteActs),
		worker.WithActivity(sa),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create installs worker: %w", err)
	}

	return wkr, nil
}
