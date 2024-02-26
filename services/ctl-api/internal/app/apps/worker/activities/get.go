package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	AppID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.App, error) {
	currentApp := app.App{}
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("Installs").
		Preload("Components").
		Preload("CreatedBy").
		First(&currentApp, "id = ?", req.AppID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
