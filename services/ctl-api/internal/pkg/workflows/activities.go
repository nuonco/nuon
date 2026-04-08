package workflows

import (
	"go.uber.org/fx"

	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitteractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	flowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type Params struct {
	fx.In

	Activities *activities.Activities
	// runner jobs
	JobActivities *jobactivities.Activities
	// workflows
	FlowActivities *flowactivities.Activities
	// shared statuses tooling
	StatusActivities *statusactivities.Activities

	// queues / signals
	QueueActs                 *queueactivities.Activities
	QueueClient               *queueclient.Client
	EmitterActs               *emitteractivities.Activities
	HandlerActs               *handleractivities.Activities
	SignalsActivities         *signalsactivities.Activities
	SignalLifecycleActivities *signal.SignalLifecycleActivities
}

type Activities struct {
	JobActivities             *jobactivities.Activities
	FlowActivities            *flowactivities.Activities
	SignalsActivities         *signalsactivities.Activities
	StatusActivities          *statusactivities.Activities
	Activities                *activities.Activities
	QueueActivities           *queueactivities.Activities
	EmitterActivities         *emitteractivities.Activities
	HandlerActivities         *handleractivities.Activities
	QueueClient               *queueclient.Client
	SignalLifecycleActivities *signal.SignalLifecycleActivities
}

func (a *Activities) AllActivities() []any {
	return []any{
		a.JobActivities,
		a.FlowActivities,
		a.Activities,
		a.SignalsActivities,
		a.StatusActivities,
		a.QueueActivities,
		a.EmitterActivities,
		a.HandlerActivities,
		a.QueueClient,
		a.SignalLifecycleActivities,
	}
}

func NewActivities(params Params) *Activities {
	return &Activities{
		Activities:                params.Activities,
		JobActivities:             params.JobActivities,
		FlowActivities:            params.FlowActivities,
		SignalsActivities:         params.SignalsActivities,
		StatusActivities:          params.StatusActivities,
		QueueActivities:           params.QueueActs,
		EmitterActivities:         params.EmitterActs,
		HandlerActivities:         params.HandlerActs,
		QueueClient:               params.QueueClient,
		SignalLifecycleActivities: params.SignalLifecycleActivities,
	}
}
