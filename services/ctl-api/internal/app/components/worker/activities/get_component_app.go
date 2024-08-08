package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentAppRequest struct {
	ComponentID string `validate:"required"`
}

// @await-gen
// @execution-timeout 5s
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
