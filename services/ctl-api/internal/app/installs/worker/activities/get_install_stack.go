package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallStackRequest struct {
	InstallID string `json:"id" validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallStack(ctx context.Context, req GetInstallStackRequest) (*app.InstallStack, error) {
	var stack app.InstallStack

	res := a.db.WithContext(ctx).
		Where(app.InstallStack{
			InstallID: req.InstallID,
		}).
		Preload("InstallStackOutputs").
		First(&stack)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install stack")
	}

	return &stack, nil
}
