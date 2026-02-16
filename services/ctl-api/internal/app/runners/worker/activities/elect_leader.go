package activities

import (
	"context"
	"fmt"
)

type ElectLeaderRequest struct {
	RunnerGroupID string `validate:"required"`
}

// @temporal-gen activity
// @by-id RunnerGroupID
func (a *Activities) ElectLeader(ctx context.Context, req ElectLeaderRequest) error {
	if err := a.helpers.ElectLeader(ctx, req.RunnerGroupID); err != nil {
		return fmt.Errorf("unable to elect leader: %w", err)
	}

	return nil
}
