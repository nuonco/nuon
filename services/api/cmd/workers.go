package cmd

import (
	"fmt"

	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build"
	buildactivities "github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteorg"
)

func (a *app) buildsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	wkflow := build.New(build.Config{Config: a.cfg.Config})
	acts := buildactivities.New(a.db)

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
	// TODO(jm): change this to the correct worker once it lands
	createWkflow := createapp.New(a.v)
	createActs := createapp.NewActivities(a.db, a.tc)

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

func (a *app) appsWorker(sa *sharedactivities.Activities) (worker.Worker, error) {
	createWkflow := createapp.New(a.v)
	createActs := createapp.NewActivities(a.db, a.tc)

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
	createActs := createorg.NewActivities(a.tc)
	deleteActs := deleteorg.NewActivities(a.tc)

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
	createActs := createinstall.NewActivities(a.db, a.tc)
	deleteActs := deleteinstall.NewActivities(a.db, a.tc)

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
