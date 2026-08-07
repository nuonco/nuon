package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) findInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Select("id", "app_config_id").
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	return &install, nil
}
