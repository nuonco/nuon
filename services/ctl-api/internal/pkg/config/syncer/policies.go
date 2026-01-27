package syncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// syncAppPolicies creates the app policies configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_policies_config.go
func (s *syncer) syncAppPolicies(ctx context.Context) error {
	if s.cfg.Policies == nil {
		return nil
	}

	policies := make([]app.AppPolicyConfig, 0, len(s.cfg.Policies.Policies))
	for _, policy := range s.cfg.Policies.Policies {
		policies = append(policies, app.AppPolicyConfig{
			AppID:       s.appID,
			AppConfigID: s.appConfigID,
			Type:        config.AppPolicyType(policy.Type),
			Engine:      config.AppPolicyEngine(policy.Engine),
			Contents:    policy.Contents,
			Components:  policy.Components,
		})
	}

	obj := app.AppPoliciesConfig{
		AppID:       s.appID,
		AppConfigID: s.appConfigID,
		Policies:    policies,
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app policies config",
			Err:         res.Error,
		}
	}

	return nil
}
