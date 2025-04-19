package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen activity
func (a *Activities) GetInstallStackOutputs(ctx context.Context, installStackID string) (*app.InstallStackOutputs, error) {
	outputs := app.InstallStackOutputs{}
	res := a.db.WithContext(ctx).
		First(&outputs, "id = ?", installStackID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	return &outputs, nil
}
