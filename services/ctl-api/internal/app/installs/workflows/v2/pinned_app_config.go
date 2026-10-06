package v2

import (
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func failIfPinnedAppConfigMoved(install *app.Install, flw *app.Workflow) error {
	stored := flw.Request
	if stored == nil || stored.RequestID == "" || stored.PinnedAppConfigID == install.AppConfigID {
		return nil
	}
	return errors.Errorf("install app config moved from %s to %s", stored.PinnedAppConfigID, install.AppConfigID)
}
