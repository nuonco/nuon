package service

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}
