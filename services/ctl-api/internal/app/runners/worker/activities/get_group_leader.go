package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	var group app.RunnerGroup
	res := a.db.WithContext(ctx).
		Select("leader_runner_id").
		First(&group, "id = ?", req.RunnerGroupID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner group: %w", res.Error)
	}

	return &GetGroupLeaderResponse{
		LeaderRunnerID: group.LeaderRunnerID,
	}, nil
}
