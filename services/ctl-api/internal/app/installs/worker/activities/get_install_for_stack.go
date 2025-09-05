package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallForStackRequest struct {
	StackID string `json:"stack_id" validate:"required"`
}

// @temporal-gen activity
// @by-id StackID
func (a *Activities) GetInstallForStack(ctx context.Context, req GetInstallForStackRequest) (*app.Install, error) {
	var stack app.InstallStack

	res := a.db.WithContext(ctx).
		Where(app.InstallStack{
			ID: req.StackID,
		}).
		First(&stack)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install stack")
	}

	return a.getInstall(ctx, stack.InstallID)
}
