package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// useStateGenV2 returns whether v2 state generation should be used for an install,
// combining the org-level feature flag with the install-level metadata override.
func (s *service) useStateGenV2(ctx context.Context, install *app.Install) (bool, error) {
	orgEnabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureStateGenV2)
	if err != nil {
		return false, fmt.Errorf("unable to check state-gen-v2 org feature: %w", err)
	}
	return pkgstate.UseStateGenV2(orgEnabled, install.Metadata), nil
}
