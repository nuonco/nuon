package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen activity
// @by-id AppID
func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.AppBranch, error) {
	currentApp := app.AppBranch{}
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("App").
		First(&currentApp, "id = ?", req.AppID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
