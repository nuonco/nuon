package statusactivities

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
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

type Params struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	MW       metrics.Writer
	Notifier FlowStatusNotifier `optional:"true"`
}

type Activities struct {
	db       *gorm.DB
	mw       metrics.Writer
	notifier FlowStatusNotifier
}

func New(params Params) *Activities {
	return &Activities{
		db:       params.DB,
		mw:       params.MW,
		notifier: params.Notifier,
	}
}
