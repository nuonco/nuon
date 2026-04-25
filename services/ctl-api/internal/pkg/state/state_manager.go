package state

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	V   *validator.Validate
}

func NewWorkflows(params Params) (*Workflows, error) {
	return &Workflows{
		cfg: params.Cfg,
		v:   params.V,
	}, nil
}

type Workflows struct {
	cfg *internal.Config
	v   *validator.Validate
}

func (w *Workflows) All() []any {
	return []any{
		w.StateManager,
	}
}

// StateManager is a long-lived per-install workflow that maintains cached state
// and only regenerates the partials that have changed.
//
// @temporal-gen-v2 workflow
// @id-template state-manager-{{.InstallID}}
// @memo type state-manager
func (w *Workflows) StateManager(ctx workflow.Context, req StateManagerRequest) error {
	sm := &stateManager{
		installID:       req.InstallID,
		state:           req.State,
		workflows:       w,
		pendingPartials: make(map[PartialName]bool),
	}
	if sm.state == nil {
		sm.state = &StateManagerState{
			LastModifiedAt: make(map[PartialName]time.Time),
		}
	}

	finished, err := sm.run(ctx)
	if err != nil {
		return err
	}
	if !finished {
		req.State = sm.state
		return workflow.NewContinueAsNewError(ctx, w.StateManager, req)
	}

	return nil
}
