package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	actionsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker"
	actionsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker/activities"
	appconfigsyncactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/appconfigsync"
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
	onboardingworker "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/worker"
	orgsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	runnersworker "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker"
	runnersactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	vcsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker"
	vcsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
)

var GeneralWorkerModule = fx.Module("worker-general",
	fx.Provide(generalactivities.New),
	fx.Provide(generalworker.NewWorkflows),
	fx.Provide(worker.AsWorker(generalworker.New)),
)

var OrgsWorkerModule = fx.Module("worker-orgs",
	fx.Provide(orgsactivities.New),
	fx.Provide(fx.Annotate(appsactivities.New, fx.ResultTags(`name:"org-trigger-activities"`))),
	fx.Provide(orgsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(orgsworker.New)),
)

var AppsWorkerModule = fx.Module("worker-apps",
	fx.Provide(appsactivities.New),
	fx.Provide(appsworker.NewWorkflows),
	fx.Provide(appbranchesactivities.New),
	fx.Provide(appconfigsyncactivities.NewActivities),
	fx.Provide(syncappconfiginstallsactivities.NewActivities),
	fx.Provide(worker.AsWorker(appsworker.New)),
)

var ComponentsWorkerModule = fx.Module("worker-components",
	fx.Provide(componentsactivities.New),
	fx.Provide(componentsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(componentsworker.New)),
)

var InstallWorkerProvidersModule = fx.Module("worker-installs-providers",
	fx.Provide(installsactivities.New),
	fx.Provide(installsworker.NewWorkflows),
	fx.Provide(installsactionsworker.NewWorkflows),
	fx.Provide(installscomponentsworker.NewWorkflows),
	fx.Provide(installssandboxworker.NewWorkflows),
	fx.Provide(installsstackworker.NewWorkflows),
)

var InstallsWorkerModule = fx.Module("worker-installs",
	fx.Provide(worker.AsWorker(installsworker.New)),
)

var InstallCronWorkerModule = fx.Module("worker-install-crons",
	fx.Provide(worker.AsWorker(installsworker.NewCronWorker)),
)

var RunnerWorkerProvidersModule = fx.Module("worker-runners-providers",
	fx.Provide(runnersactivities.New),
	fx.Provide(runnersworker.NewWorkflows),
)

var RunnersWorkerModule = fx.Module("worker-runners",
	fx.Provide(worker.AsWorker(runnersworker.New)),
)

var RunnerHealthcheckCronWorkerModule = fx.Module("worker-runner-healthcheck-crons",
	fx.Provide(worker.AsWorker(runnersworker.NewHealthcheckCronWorker)),
)

var ActionsWorkerModule = fx.Module("worker-actions",
	fx.Provide(actionsactivities.New),
	fx.Provide(actionsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(actionsworker.New)),
)

var OnboardingsWorkerModule = fx.Module("worker-onboardings",
	fx.Provide(worker.AsWorker(onboardingworker.New)),
)

var VCSWorkerModule = fx.Module("worker-vcs",
	fx.Provide(func(h *vcshelpers.Helpers) vcsactivities.GithubClient { return h }),
	fx.Provide(vcsactivities.New),
	fx.Provide(worker.AsWorker(vcsworker.New)),
)
