package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type stepGroup struct {
	idx int
}

func newStepGroup() *stepGroup {
	return &stepGroup{}
}

func (s *stepGroup) nextGroup() {
	s.idx++
}

func (s *stepGroup) installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, sig signal.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	opts = append(opts, WithGroupIdx(s.idx))
	return installSignalStep(ctx, installID, name, metadata, sig, planOnly, opts...)
}
