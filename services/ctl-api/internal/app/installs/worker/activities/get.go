package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	return a.helpers.GetInstallByID(ctx, installID)
}
