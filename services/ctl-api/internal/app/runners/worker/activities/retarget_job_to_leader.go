package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type RetargetJobToLeaderRequest struct {
	JobID    string `validate:"required"`
	RunnerID string `validate:"required"`
}

type RetargetJobToLeaderResponse struct {
	NoLeader       bool
	Retargeted     bool
	LeaderRunnerID string
}

// @temporal-gen activity
// @by-id RunnerID
func (a *Activities) RetargetJobToLeader(ctx context.Context, req RetargetJobToLeaderRequest) (*RetargetJobToLeaderResponse, error) {
	resp := &RetargetJobToLeaderResponse{}

	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Look up the runner.
		var runner app.Runner
		if res := tx.First(&runner, "id = ? AND deleted_at = 0", req.RunnerID); res.Error != nil {
			return fmt.Errorf("unable to get runner: %w", res.Error)
		}

		// Find the leader directly instead of loading all runners in the group.
		var leader app.Runner
		err := tx.Where("runner_group_id = ? AND leader = true AND deleted_at = 0", runner.RunnerGroupID).
			First(&leader).Error
		if err != nil {
			resp.NoLeader = true
			return nil
		}

		// This runner is already the leader.
		if leader.ID == req.RunnerID {
			return nil
		}

		// Retarget the job to the leader runner.
		if res := tx.Model(&app.RunnerJob{}).
			Where("id = ? AND deleted_at = 0", req.JobID).
			Update("runner_id", leader.ID); res.Error != nil {
			return fmt.Errorf("unable to retarget job to leader: %w", res.Error)
		}

		resp.Retargeted = true
		resp.LeaderRunnerID = leader.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
