package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnerRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetRunner(ctx context.Context, req GetRunnerRequest) (*app.Runner, error) {
	runner := app.Runner{}
	res := a.db.WithContext(ctx).
		First(&runner, "id = ?", req.ID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
