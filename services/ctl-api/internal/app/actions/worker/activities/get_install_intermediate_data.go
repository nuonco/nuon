package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallIntermediateDataRequest struct {
	InstallID string `json:"install_id"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetInstallIntermediateData(ctx context.Context, req *GetInstallIntermediateDataRequest) (*app.InstallIntermediateData, error) {
	var intermediateData app.InstallIntermediateData

	res := a.db.WithContext(ctx).
		Order("created_at desc").
		First(&intermediateData, "install_id = ?", req.InstallID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to fetch install intermediate data")
	}

	return &intermediateData, nil
}
