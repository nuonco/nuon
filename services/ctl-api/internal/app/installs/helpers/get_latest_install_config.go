package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) GetLatestInstallConfig(
	ctx context.Context, installID string) (*app.InstallConfig, error) {
	installConfig := app.InstallConfig{}
	resp := s.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&installConfig)
	if resp.Error != nil {
		if resp.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(resp.Error, "error fetching install config")
	}

	return &installConfig, nil
}
