package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallInputsStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallInputsState(ctx context.Context, req GetInstallInputsStateRequest) (*app.InstallInputs, error) {
	var inps app.InstallInputs
	res := a.db.WithContext(ctx).
		Where(app.InstallInputs{
			InstallID: req.InstallID,
		}).
		Order("created_at desc").
		First(&inps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install inputs")
	}

	return &inps, nil
}
