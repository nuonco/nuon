package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateVCSConnectionStatusRequest struct {
	VCSConnectionID string         `validate:"required"`
	Status          app.Status     `validate:"required"`
	Description     string         `json:"description"`
	Metadata        map[string]any `json:"metadata"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateVCSConnectionStatus(ctx context.Context, req UpdateVCSConnectionStatusRequest) error {
	status := &app.CompositeStatus{
		CreatedAtTS:            time.Now().Unix(),
		Status:                 req.Status,
		StatusHumanDescription: req.Description,
		Metadata:               req.Metadata,
	}

	res := a.db.WithContext(ctx).
		Model(&app.VCSConnection{}).
		Where("id = ?", req.VCSConnectionID).
		Update("status", status)
	if res.Error != nil {
		return fmt.Errorf("unable to update vcs connection status: %w", res.Error)
	}

	return nil
}
