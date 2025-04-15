package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

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
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", req.ID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("not found", "not found", res.Error, "")
		}

		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
