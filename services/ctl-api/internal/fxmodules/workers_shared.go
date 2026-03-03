package fxmodules

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/interceptors"
	cctxinterceptor "github.com/nuonco/nuon/services/ctl-api/internal/interceptors/cctx"
	metricsinterceptor "github.com/nuonco/nuon/services/ctl-api/internal/interceptors/metrics"
	validateinterceptor "github.com/nuonco/nuon/services/ctl-api/internal/interceptors/validate"
	queue "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowsflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow"
	flowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"

	// Register app branch run signal in catalog (blank import triggers init())
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/run"
)

// WorkerInterceptorsModule provides interceptors for temporal workers.
var WorkerInterceptorsModule = fx.Module("worker-interceptors",
	fx.Provide(interceptors.AsInterceptor(cctxinterceptor.New)),
	fx.Provide(interceptors.AsInterceptor(metricsinterceptor.New)),
	fx.Provide(interceptors.AsInterceptor(validateinterceptor.New)),
)

// SharedWorkflowsModule provides shared workflow activities and workflows
// used across multiple worker namespaces.
var SharedWorkflowsModule = fx.Module("shared-workflows",
	fx.Provide(jobactivities.New),
	fx.Provide(flowactivities.New),
	fx.Provide(signalsactivities.New),
	fx.Provide(queueactivities.New),
	fx.Provide(handleractivities.New),
	fx.Provide(statusactivities.New),
	fx.Provide(activities.New),

	// workflows
	fx.Provide(job.New),
	fx.Provide(workflowsflow.New),
	fx.Provide(workflows.NewActivities),
	fx.Provide(workflows.NewWorkflows),
	fx.Provide(queue.NewWorkflows),
	fx.Provide(handler.NewWorkflows),
)
