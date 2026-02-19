package activities

import (
	"context"
	"fmt"
)

type SetLeaderRequest struct {
	RunnerGroupID string `validate:"required"`
	RunnerID      string `validate:"required"`
}

type SetLeaderResponse struct {
	OldLeaderID string
	NewLeaderID string
}

// @temporal-gen activity
// @by-id RunnerGroupID
func (a *Activities) SetLeader(ctx context.Context, req SetLeaderRequest) (*SetLeaderResponse, error) {
	result, err := a.helpers.SetLeader(ctx, req.RunnerGroupID, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to set leader: %w", err)
	}

	return &SetLeaderResponse{
		OldLeaderID: result.OldLeaderID,
		NewLeaderID: result.NewLeaderID,
	}, nil
}
