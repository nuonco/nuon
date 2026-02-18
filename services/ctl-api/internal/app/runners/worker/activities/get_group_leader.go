package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetGroupLeaderRequest struct {
	RunnerGroupID string `validate:"required"`
}

type GetGroupLeaderResponse struct {
	LeaderRunnerID *string
}

// @temporal-gen activity
// @by-id RunnerGroupID
func (a *Activities) GetGroupLeader(ctx context.Context, req GetGroupLeaderRequest) (*GetGroupLeaderResponse, error) {
	var leader app.Runner
	err := a.db.WithContext(ctx).
		Where("runner_group_id = ? AND leader = true AND deleted_at = 0", req.RunnerGroupID).
		First(&leader).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &GetGroupLeaderResponse{LeaderRunnerID: nil}, nil
		}
		return nil, fmt.Errorf("unable to get group leader: %w", err)
	}

	return &GetGroupLeaderResponse{
		LeaderRunnerID: &leader.ID,
	}, nil
}
