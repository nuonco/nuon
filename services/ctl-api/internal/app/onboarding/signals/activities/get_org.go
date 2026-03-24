package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field orgID
func (a *Activities) getOrgByID(ctx context.Context, orgID string) (*app.Org, error) {
	var org app.Org
	if err := a.db.WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}
	return &org, nil
}
