package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type SetLeaderResult struct {
	OldLeaderID string
	NewLeaderID string
}

func (s *Helpers) SetLeader(ctx context.Context, groupID, runnerID string) (*SetLeaderResult, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	var group app.RunnerGroup
	if err := tx.Raw("SELECT * FROM runner_groups WHERE id = ? AND deleted_at = 0 FOR UPDATE", groupID).Scan(&group).Error; err != nil {
		return nil, fmt.Errorf("unable to lock runner group: %w", err)
	}
	if group.ID == "" {
		return nil, fmt.Errorf("runner group not found: %s", groupID)
	}

	// Verify the requested runner exists, is active, and untainted.
	var runner app.Runner
	err := tx.Where("id = ? AND runner_group_id = ? AND deleted_at = 0", runnerID, groupID).First(&runner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("runner %s not found in group %s", runnerID, groupID)
		}
		return nil, fmt.Errorf("unable to query runner: %w", err)
	}
	if runner.Status != app.RunnerStatusActive {
		return nil, fmt.Errorf("runner %s is not active (status: %s)", runnerID, runner.Status)
	}

	// Find the current leader.
	var oldLeader app.Runner
	var oldLeaderID string
	if err := tx.Where("runner_group_id = ? AND leader = true AND deleted_at = 0", groupID).First(&oldLeader).Error; err == nil {
		oldLeaderID = oldLeader.ID
	}

	// Already the leader — no-op.
	if oldLeaderID == runnerID {
		return &SetLeaderResult{
			OldLeaderID: oldLeaderID,
			NewLeaderID: runnerID,
		}, nil
	}

	// Clear all leader flags in the group.
	if err := tx.Model(&app.Runner{}).
		Where("runner_group_id = ? AND deleted_at = 0", groupID).
		Update("leader", false).Error; err != nil {
		return nil, fmt.Errorf("unable to clear leader flags: %w", err)
	}

	// Set the requested runner as leader.
	if err := tx.Model(&app.Runner{}).
		Where("id = ? AND deleted_at = 0", runnerID).
		Update("leader", true).Error; err != nil {
		return nil, fmt.Errorf("unable to set leader: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("unable to commit set leader: %w", err)
	}

	return &SetLeaderResult{
		OldLeaderID: oldLeaderID,
		NewLeaderID: runnerID,
	}, nil
}
