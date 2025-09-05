package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type GetJobRequest struct {
	ID string `validate:"required"`
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) PkgWorkflowsJobGetJob(ctx context.Context, req *GetJobRequest) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := a.db.WithContext(ctx).
	Scopes(scopes.WithDisableViews).
		First(&job, "id = ?", req.ID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &job, nil
}
