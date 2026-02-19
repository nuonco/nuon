package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type ElectLeaderResult struct {
	OldLeaderID string
	NewLeaderID string
}

func (s *Helpers) ElectLeader(ctx context.Context, groupID string) (*ElectLeaderResult, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	var group app.RunnerGroup
	if err := tx.Raw("SELECT * FROM runner_groups WHERE id = ? AND deleted_at = 0 FOR UPDATE", groupID).Scan(&group).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("unable to lock runner group: %w", err)
	}
	if group.ID == "" {
		tx.Rollback()
		return nil, fmt.Errorf("runner group not found: %s", groupID)
	}

	// Find the current leader in this group.
	var oldLeader app.Runner
	var oldLeaderID string
	if err := tx.Where("runner_group_id = ? AND leader = true AND deleted_at = 0", groupID).First(&oldLeader).Error; err == nil {
		oldLeaderID = oldLeader.ID
	}

	// Clear all leader flags in the group.
	if err := tx.Model(&app.Runner{}).
		Where("runner_group_id = ? AND deleted_at = 0", groupID).
		Update("leader", false).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("unable to clear leader flags: %w", err)
	}

	// Elect the oldest active, untainted runner as leader.
	var leader app.Runner
	err := tx.
		Where("runner_group_id = ? AND status = ? AND tainted = false AND deleted_at = 0", groupID, app.RunnerStatusActive).
		Order("created_at ASC").
		First(&leader).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to query active runners: %w", err)
		}
		// No active runners — leader stays cleared.
	} else {
		if err := tx.Model(&app.Runner{}).
			Where("id = ? AND deleted_at = 0", leader.ID).
			Update("leader", true).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("unable to set leader: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("unable to commit leader election: %w", err)
	}

	result := &ElectLeaderResult{
		OldLeaderID: oldLeaderID,
	}
	if err == nil {
		result.NewLeaderID = leader.ID
	}

	return result, nil
}
