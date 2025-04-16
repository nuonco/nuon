package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallStackRequest struct {
	InstallID string `json:"id"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallStack(ctx context.Context, req GetInstallStackRequest) (*app.InstallStack, error) {
	var stack app.InstallStack

	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{
			InstallID: req.InstallID,
		}).
		First(&stack); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install stack")
	}

	return &stack, nil
}
