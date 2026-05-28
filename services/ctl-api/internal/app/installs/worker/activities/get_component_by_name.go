package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetComponentByNameRequest struct {
	InstallID     string `validate:"required"`
	ComponentName string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetComponentByName(ctx context.Context, req GetComponentByNameRequest) (*app.Component, error) {
	// Get the install to find its app
	var install app.Install
	res := a.db.WithContext(ctx).First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	var component app.Component
	res = a.db.WithContext(ctx).
		Where(app.Component{AppID: install.AppID, Name: req.ComponentName}).
		First(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find component %s: %w", req.ComponentName, res.Error)
	}

	return &component, nil
}
