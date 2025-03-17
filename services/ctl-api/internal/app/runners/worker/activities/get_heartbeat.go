package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetHeartBeatRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetHeartBeat(ctx context.Context, req GetHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	runner := app.RunnerHeartBeat{}
	res := a.chDB.WithContext(ctx).
		First(&runner, "id = ?", req.ID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
