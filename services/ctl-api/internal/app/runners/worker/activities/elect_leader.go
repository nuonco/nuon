package activities

import (
	"context"
	"fmt"
)

type ElectLeaderRequest struct {
	RunnerGroupID string `validate:"required"`
}

type ElectLeaderResponse struct {
	OldLeaderID string
	NewLeaderID string
}

// @temporal-gen activity
// @by-id RunnerGroupID
func (a *Activities) ElectLeader(ctx context.Context, req ElectLeaderRequest) (*ElectLeaderResponse, error) {
	result, err := a.helpers.ElectLeader(ctx, req.RunnerGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to elect leader: %w", err)
	}

	return &ElectLeaderResponse{
		OldLeaderID: result.OldLeaderID,
		NewLeaderID: result.NewLeaderID,
	}, nil
}
