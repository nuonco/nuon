package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type CreateSandboxInstallStackVersionRunRequest struct {
	StackVersionID string `validate:"required"`

	Data map[string]string `validate:"required"`
}

// @temporal-gen activity
// @by-id StackVersionID
func (a *Activities) CreateSandboxInstallStackVersionRun(ctx context.Context, req *CreateSandboxInstallStackVersionRunRequest) (*app.InstallStackVersionRun, error) {
	versionRun := app.InstallStackVersionRun{
		InstallStackVersionID: req.StackVersionID,
		Data:                  generics.ToHstore(req.Data),
	}
	res := a.db.WithContext(ctx).
		Create(&versionRun)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to create install stack version id")
	}

	return &versionRun, nil
}
