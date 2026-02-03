package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	actionsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker"
	actionsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/worker/activities"
	appbranchesworker "github.com/nuonco/nuon/services/ctl-api/internal/app/app-branches/worker"
	appbranchesactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/app-branches/worker/activities"
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
	installstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	orgsworker "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker"
	orgsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	releasesworker "github.com/nuonco/nuon/services/ctl-api/internal/app/releases/worker"
	releasesactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/releases/worker/activities"
	runnersworker "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker"
	runnersactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
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
	fx.Provide(orgsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(orgsworker.New)),
)

// AppsWorkerModule provides the apps namespace worker.
var AppsWorkerModule = fx.Module("worker-apps",
	fx.Provide(appsactivities.New),
	fx.Provide(appsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(appsworker.New)),
)

// AppBranchesWorkerModule provides the app-branches namespace worker.
var AppBranchesWorkerModule = fx.Module("worker-app-branches",
	fx.Provide(appbranchesactivities.New),
	fx.Provide(appbranchesworker.NewWorkflows),
	fx.Provide(worker.AsWorker(appbranchesworker.New)),
)

// ComponentsWorkerModule provides the components namespace worker.
var ComponentsWorkerModule = fx.Module("worker-components",
	fx.Provide(componentsactivities.New),
	fx.Provide(componentsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(componentsworker.New)),
)

// InstallsWorkerModule provides the installs namespace worker.
var InstallsWorkerModule = fx.Module("worker-installs",
	fx.Provide(installsactivities.New),
	fx.Provide(installsworker.NewWorkflows),
	fx.Provide(installsactionsworker.NewWorkflows),
	fx.Provide(installscomponentsworker.NewWorkflows),
	fx.Provide(installssandboxworker.NewWorkflows),
	fx.Provide(installsstackworker.NewWorkflows),
	fx.Provide(installstate.New),
	fx.Provide(worker.AsWorker(installsworker.New)),
)

// ReleasesWorkerModule provides the releases namespace worker.
var ReleasesWorkerModule = fx.Module("worker-releases",
	fx.Provide(releasesactivities.New),
	fx.Provide(releasesworker.NewWorkflows),
	fx.Provide(worker.AsWorker(releasesworker.New)),
)

// RunnersWorkerModule provides the runners namespace worker.
var RunnersWorkerModule = fx.Module("worker-runners",
	fx.Provide(runnersactivities.New),
	fx.Provide(runnersworker.NewWorkflows),
	fx.Provide(worker.AsWorker(runnersworker.New)),
)

// ActionsWorkerModule provides the actions namespace worker.
var ActionsWorkerModule = fx.Module("worker-actions",
	fx.Provide(actionsactivities.New),
	fx.Provide(actionsworker.NewWorkflows),
	fx.Provide(worker.AsWorker(actionsworker.New)),
)
