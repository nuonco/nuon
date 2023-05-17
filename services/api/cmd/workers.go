package cmd

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/build"
	buildactivities "github.com/powertoolsdev/mono/services/api/internal/jobs/build/activities"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createapp"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/createorg"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteinstall"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/deleteorg"
)

func (a *app) loadWorkers(domain string) ([]worker.Worker, error) {
	workers := make([]worker.Worker, 0)
	if domain == "apps" || domain == "all" {
		wkr, err := a.appsWorker()
		if err != nil {
			return nil, fmt.Errorf("unable to load apps worker: %w", err)
		}
		workers = append(workers, wkr)
	}

	if domain == "builds" || domain == "all" {
		wkr, err := a.buildsWorker()
		if err != nil {
			return nil, fmt.Errorf("unable to load builds worker: %w", err)
		}
		workers = append(workers, wkr)
	}

	if domain == "deploys" || domain == "all" {
		wkr, err := a.deploysWorker()
		if err != nil {
			return nil, fmt.Errorf("unable to load deploys worker: %w", err)
		}
		workers = append(workers, wkr)
	}

	if domain == "installs" || domain == "all" {
		wkr, err := a.installsWorker()
		if err != nil {
			return nil, fmt.Errorf("unable to load installs worker: %w", err)
		}
		workers = append(workers, wkr)
	}

	if domain == "orgs" || domain == "all" {
		wkr, err := a.orgsWorker()
		if err != nil {
			return nil, fmt.Errorf("unable to load orgs worker: %w", err)
		}
		workers = append(workers, wkr)
	}

	return workers, nil
}

func (a *app) buildsWorker() (worker.Worker, error) {
	wkflow := build.New(build.Config{Config: a.cfg.Config})
	acts := buildactivities.New(a.db)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithNamespace("builds"),

		worker.WithWorkflow(wkflow.Build),

		worker.WithActivity(acts),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create builds worker: %w", err)
	}

	return wkr, nil
}

func (a *app) deploysWorker() (worker.Worker, error) {
	// TODO(jm): change this to the correct worker once it lands
	createWkflow := createapp.New(a.v)
	createActs := createapp.NewActivities(a.db, a.tc)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithWorkflow(createWkflow.CreateApp),
		worker.WithNamespace("apps"),
		worker.WithActivity(createActs),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create apps worker: %w", err)
	}

	return wkr, nil
}

func (a *app) appsWorker() (worker.Worker, error) {
	createWkflow := createapp.New(a.v)
	createActs := createapp.NewActivities(a.db, a.tc)

	wkr, err := worker.New(a.v, worker.WithConfig(&a.cfg.Config),
		worker.WithWorkflow(createWkflow.CreateApp),
		worker.WithNamespace("apps"),
		worker.WithActivity(createActs),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create apps worker: %w", err)
	}

	return wkr, nil
}

func (a *app) orgsWorker() (worker.Worker, error) {
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
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create orgs worker: %w", err)
	}

	return wkr, nil
}

func (a *app) installsWorker() (worker.Worker, error) {
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
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create installs worker: %w", err)
	}

	return wkr, nil
}
