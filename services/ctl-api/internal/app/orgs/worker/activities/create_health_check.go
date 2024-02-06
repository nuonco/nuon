package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateHealthCheckRequest struct {
	OrgID string `validate:"required"`
}

func (a *Activities) CreateHealthCheck(ctx context.Context, req CreateHealthCheckRequest) (*app.OrgHealthCheck, error) {
	org, err := a.getOrg(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}

	healthCheck := app.OrgHealthCheck{
		CreatedByID:       org.CreatedByID,
		OrgID:             req.OrgID,
		Status:            app.OrgHealthCheckStatusInProgress,
		StatusDescription: "health check in-progress",
	}
	res := a.db.WithContext(ctx).Create(&healthCheck)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create health check: %w", res.Error)
	}

	return &healthCheck, nil
}
