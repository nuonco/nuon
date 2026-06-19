package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type CountInstallStackVersionRunsRequest struct {
	VersionID string `json:"version_id" validate:"required"`
}

type CountInstallStackVersionRunsResponse struct {
	Count int64 `json:"count"`
}

// @temporal-gen-v2 activity
func (a *Activities) CountInstallStackVersionRuns(ctx context.Context, req CountInstallStackVersionRunsRequest) (*CountInstallStackVersionRunsResponse, error) {
	var count int64
	if res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersionRun{}).
		Where(app.InstallStackVersionRun{InstallStackVersionID: req.VersionID}).
		Count(&count); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to count runs")
	}
	return &CountInstallStackVersionRunsResponse{Count: count}, nil
}
