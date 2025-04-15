package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetAWSCloudFormationStackRequest struct {
	InstallID string `json:"id"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetAWSCloudFormationStack(ctx context.Context, req GetAWSCloudFormationStackRequest) (*app.InstallAWSCloudFormationStack, error) {
	var stack app.InstallAWSCloudFormationStack

	if res := a.db.WithContext(ctx).
		Where(app.InstallAWSCloudFormationStack{
			InstallID: req.InstallID,
		}).
		First(&stack); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get cloudformation stack")
	}

	return &stack, nil
}
