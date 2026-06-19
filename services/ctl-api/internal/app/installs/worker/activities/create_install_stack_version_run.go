package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type CreateInstallStackVersionRunRequest struct {
	InstallStackVersionID string        `json:"install_stack_version_id" validate:"required"`
	Data                  pgtype.Hstore `json:"data"`
}

type CreateInstallStackVersionRunResponse struct {
	ID string `json:"id"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateInstallStackVersionRun(ctx context.Context, req CreateInstallStackVersionRunRequest) (*CreateInstallStackVersionRunResponse, error) {
	run := app.InstallStackVersionRun{
		InstallStackVersionID: req.InstallStackVersionID,
		Data:                  req.Data,
	}
	if res := a.db.WithContext(ctx).Create(&run); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to create install stack version run")
	}
	return &CreateInstallStackVersionRunResponse{ID: run.ID}, nil
}
