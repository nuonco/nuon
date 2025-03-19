package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRunnersRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetRunners(ctx context.Context, req GetRunnersRequest) ([]*app.Runner, error) {
	var runners []*app.Runner

	res := a.db.WithContext(ctx).
		Where(app.Runner{
			OrgID: req.ID,
		}).
		Find(&runners)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return runners, nil
}
