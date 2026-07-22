package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	actionsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker"
	actionsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker/activities"
	appbranchesactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	syncappconfiginstallsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/syncappconfiginstalls"
	appsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker"
	appsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	componentsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker"
	componentsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	generalworker "github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker"
	generalactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	installsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker"
	installsactionsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
	installsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	installscomponentsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/components"
	installssandboxworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/sandbox"
	installsstackworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/stack"
	installsstateworker "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	onboardingworker "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/worker"
	orgsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnersworker "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker"
	runnersactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	vcsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker"
	vcsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
)

// GeneralWorkerModule provides the general namespace worker.
var GeneralWorkerModule = fx.Module("worker-general",
	fx.Provide(generalactivities.New),
	fx.Provide(generalworker.NewWorkflows),
	fx.Provide(worker.AsWorker(generalworker.New)),
)

// OrgsWorkerModule provides the orgs namespace worker.
var OrgsWorkerModule = fx.Module("worker-orgs",
	fx.Provide(orgsactivities.New),
	fx.Provide(fx.Annotate(appsactivities.New, fx.ResultTags(`name:"org-automation-activities"`))),
	fx.Provide(orgsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(orgsworker.New)),
)

// AppsWorkerModule provides the apps namespace worker.
var AppsWorkerModule = fx.Module("worker-apps",
	fx.Provide(appsactivities.New),
	fx.Provide(appsworker.NewWorkflows),
	fx.Provide(appbranchesactivities.New),
	fx.Provide(syncappconfiginstallsactivities.NewActivities),
	fx.Provide(worker.AsWorker(appsworker.New)),
)

// ComponentsWorkerModule provides the components namespace worker.
var ComponentsWorkerModule = fx.Module("worker-components",
	fx.Provide(componentsactivities.New),
	fx.Provide(componentsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(componentsworker.New)),
)

// InstallWorkerProvidersModule provides the install workflow/activity
// constructors shared by the installs worker and the install crons worker.
// Kept separate so either worker can run standalone without double-providing.
var InstallWorkerProvidersModule = fx.Module("worker-installs-providers",
	fx.Provide(installsactivities.New),
	fx.Provide(installsworker.NewWorkflows),
	fx.Provide(installsactionsworker.NewWorkflows),
	fx.Provide(installscomponentsworker.NewWorkflows),
	fx.Provide(installssandboxworker.NewWorkflows),
	fx.Provide(installsstackworker.NewWorkflows),
	fx.Provide(installsstateworker.New),
)

// InstallsWorkerModule provides the installs namespace worker (api task queue).
var InstallsWorkerModule = fx.Module("worker-installs",
	fx.Provide(worker.AsWorker(installsworker.New)),
)

// InstallCronWorkerModule provides the install crons worker (install-crons task queue).
var InstallCronWorkerModule = fx.Module("worker-install-crons",
	fx.Provide(worker.AsWorker(installsworker.NewCronWorker)),
)

// RunnerWorkerProvidersModule provides the runner workflow/activity constructors
// shared by the runners worker and the runner healthcheck crons worker.
var RunnerWorkerProvidersModule = fx.Module("worker-runners-providers",
	fx.Provide(runnersactivities.New),
	fx.Provide(runnersworker.NewWorkflows),
)

// RunnersWorkerModule provides the runners namespace worker (api task queue).
var RunnersWorkerModule = fx.Module("worker-runners",
	fx.Provide(worker.AsWorker(runnersworker.New)),
)

// RunnerHealthcheckCronWorkerModule provides the runner healthcheck crons worker.
var RunnerHealthcheckCronWorkerModule = fx.Module("worker-runner-healthcheck-crons",
	fx.Provide(worker.AsWorker(runnersworker.NewHealthcheckCronWorker)),
)

// ActionsWorkerModule provides the actions namespace worker.
var ActionsWorkerModule = fx.Module("worker-actions",
	fx.Provide(actionsactivities.New),
	fx.Provide(actionsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(actionsworker.New)),
)

// OnboardingsWorkerModule provides the onboardings namespace worker.
var OnboardingsWorkerModule = fx.Module("worker-onboardings",
	fx.Provide(worker.AsWorker(onboardingworker.New)),
)

// VCSWorkerModule provides the vcs namespace worker.
var VCSWorkerModule = fx.Module("worker-vcs",
	fx.Provide(func(h *vcshelpers.Helpers) vcsactivities.GithubClient { return h }),
	fx.Provide(vcsactivities.New),
	fx.Provide(worker.AsWorker(vcsworker.New)),
)
