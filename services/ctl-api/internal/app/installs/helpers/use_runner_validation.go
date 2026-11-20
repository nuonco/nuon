package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// ValidateUseRunnerFalse checks the preconditions for setting an install's
// desired use_runner to false: an AWS cloud connection and the org's
// runnerless-installs feature. Shared by install creation and
// UpdateInstallConfig so both reject the same way for the same reasons.
func (s *Helpers) ValidateUseRunnerFalse(ctx context.Context, cloudConnectionID string, isAWS bool) error {
	if cloudConnectionID == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("use_runner=false requires cloud_connection_id"),
			Description: "use_runner=false requires a cloud_connection_id so control-plane workers can assume the install stack's roles",
		}
	}
	if !isAWS {
		return stderr.ErrUser{
			Err:         fmt.Errorf("use_runner=false is only supported for AWS installs"),
			Description: "use_runner=false is only supported for AWS installs",
		}
	}
	runnerlessEnabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureRunnerlessInstalls)
	if err != nil {
		return fmt.Errorf("check runnerless-installs feature: %w", err)
	}
	if !runnerlessEnabled {
		return stderr.ErrUser{
			Err:         fmt.Errorf("runnerless-installs feature is not enabled for this org"),
			Description: "use_runner=false requires the runnerless-installs feature to be enabled for this org",
		}
	}
	return nil
}
