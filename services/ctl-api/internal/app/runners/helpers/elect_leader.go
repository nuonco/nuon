package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) ElectLeader(ctx context.Context, groupID string) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("unable to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	var group app.RunnerGroup
	if err := tx.Raw("SELECT * FROM runner_groups WHERE id = ? AND deleted_at = 0 FOR UPDATE", groupID).Scan(&group).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to lock runner group: %w", err)
	}
	if group.ID == "" {
		tx.Rollback()
		return fmt.Errorf("runner group not found: %s", groupID)
	}

	var leader app.Runner
	err := tx.
		Where("runner_group_id = ? AND status = ? AND deleted_at = 0", groupID, app.RunnerStatusActive).
		Order("created_at ASC").
		First(&leader).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to query active runners: %w", err)
		}
		// No active runners — clear the leader.
		if err := tx.Model(&app.RunnerGroup{ID: groupID}).Update("leader_runner_id", nil).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to clear leader: %w", err)
		}
	} else {
		if err := tx.Model(&app.RunnerGroup{ID: groupID}).Update("leader_runner_id", leader.ID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to set leader: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("unable to commit leader election: %w", err)
	}

	return nil
}
