package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveStateRequest struct {
	State *state.State `validate:"required"`

	InstallID       string `validate:"required"`
	TriggeredByID   string `validate:"required"`
	TriggeredByType string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SaveState(ctx context.Context, req *SaveStateRequest) (*app.InstallState, error) {
	obj := &app.InstallState{
		InstallID:       req.InstallID,
		TriggeredByID:   req.TriggeredByID,
		TriggeredByType: req.TriggeredByType,
		State:           req.State,
	}

	res := a.db.WithContext(ctx).
		Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create install state")
	}
	return obj, nil
}
