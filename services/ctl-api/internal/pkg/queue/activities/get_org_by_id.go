package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @task-queue "api"
// @start-to-close-timeout 10s
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field orgID
func (a *Activities) getOrgByID(ctx context.Context, orgID string) (*app.Org, error) {
	var org app.Org
	if res := a.db.WithContext(ctx).
		Where(app.Org{ID: orgID}).
		First(&org); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get org")
	}
	return &org, nil
}
