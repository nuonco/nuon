package statusactivities

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// FlowStatusNotifier is an optional hook invoked after a flow status update
// is persisted. Implemented outside this package (see
// pkg/flow/signals/workflowstepawaitingretry) because the queue client can't
// be imported here without an import cycle (queue/handler imports this
// package). Implementations must be best-effort: they may not fail the
// status update, so the method returns nothing.
type FlowStatusNotifier interface {
	FlowStatusUpdated(ctx context.Context, req UpdateStatusRequest)
}

type WorkflowMetrics interface {
	SignalStatusUpdated(context.Context, app.QueueSignal, app.Status)
	StepStatusUpdated(context.Context, app.WorkflowStep, app.CompositeStatus)
}

type Params struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	MW       metrics.Writer
	Notifier FlowStatusNotifier `optional:"true"`
	L        *zap.Logger        `optional:"true"`
	Counters WorkflowMetrics    `optional:"true"`
}

type Activities struct {
	db       *gorm.DB
	mw       metrics.Writer
	notifier FlowStatusNotifier
	l        *zap.Logger
	counters WorkflowMetrics
}

func New(params Params) *Activities {
	l := params.L
	if l == nil {
		l = zap.NewNop()
	}
	return &Activities{
		db:       params.DB,
		mw:       params.MW,
		l:        l,
		notifier: params.Notifier,
		counters: params.Counters,
	}
}
