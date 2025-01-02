package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"
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
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, temporal.NewNonRetryableApplicationError("not found", "not found", res.Error, "")
		}

		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
