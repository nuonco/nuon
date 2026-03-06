package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetComponentAppRequest struct {
	ComponentID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ComponentID
func (a *Activities) GetComponentApp(ctx context.Context, req *GetComponentAppRequest) (*app.App, error) {
	cmp := app.Component{}
	res := a.db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").
		First(&cmp, "id = ?", req.ComponentID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &cmp.App, nil
}
