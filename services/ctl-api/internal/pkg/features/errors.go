package features

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ErrFeatureNotEnabled returns a standard user error when a required feature is not enabled
func ErrFeatureNotEnabled(feature app.OrgFeature) stderr.ErrUser {
	return stderr.ErrUser{
		Err:         fmt.Errorf("feature not enabled: %s", feature),
		Description: fmt.Sprintf("This operation requires the '%s' feature to be enabled for your organization", feature),
		Code:        "feature_not_enabled",
	}
}
