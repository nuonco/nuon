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
		// Look up the runner with its group and sibling runners to find the leader.
		var runner app.Runner
		if res := tx.Preload("RunnerGroup.Runners").First(&runner, "id = ? AND deleted_at = 0", req.RunnerID); res.Error != nil {
			return fmt.Errorf("unable to get runner: %w", res.Error)
		}

		leader := runner.RunnerGroup.ActiveRunner()
		if leader == nil || !leader.Leader {
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
