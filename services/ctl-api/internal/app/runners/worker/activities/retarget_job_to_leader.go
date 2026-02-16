package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
		// Look up the runner to get its group ID.
		var runner app.Runner
		if res := tx.Select("runner_group_id").First(&runner, "id = ? AND deleted_at = 0", req.RunnerID); res.Error != nil {
			return fmt.Errorf("unable to get runner: %w", res.Error)
		}

		// Load fresh leader_runner_id from the runner group.
		var group app.RunnerGroup
		if res := tx.Select("leader_runner_id").First(&group, "id = ? AND deleted_at = 0", runner.RunnerGroupID); res.Error != nil {
			return fmt.Errorf("unable to get runner group: %w", res.Error)
		}

		// No leader elected yet.
		if group.LeaderRunnerID == nil {
			resp.NoLeader = true
			return nil
		}

		// This runner is already the leader.
		if *group.LeaderRunnerID == req.RunnerID {
			resp.Retargeted = false
			return nil
		}

		// Retarget the job to the leader runner.
		if res := tx.Model(&app.RunnerJob{}).
			Where("id = ? AND deleted_at = 0", req.JobID).
			Update("runner_id", *group.LeaderRunnerID); res.Error != nil {
			return fmt.Errorf("unable to retarget job to leader: %w", res.Error)
		}

		resp.Retargeted = true
		resp.LeaderRunnerID = *group.LeaderRunnerID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
