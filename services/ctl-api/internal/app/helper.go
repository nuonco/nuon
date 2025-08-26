package app

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func MarkInstallStateStale(tx *gorm.DB, installID string) error {
	var is InstallState

	err := tx.Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&is).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // No record to update
		}
		return err
	}

	return tx.Model(&is).Update("stale_at", time.Now()).Error
}
