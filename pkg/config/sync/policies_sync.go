package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) policyToRequest(policy config.AppPolicy) *models.ServiceAppPolicyConfig {
	return &models.ServiceAppPolicyConfig{
		Contents: generics.ToPtr(policy.Contents),
		Type:     generics.ToPtr(policy.Type),
	}
}

func (s sync) getAppPoliciesRequest() *models.ServiceCreateAppPoliciesConfigRequest {
	req := &models.ServiceCreateAppPoliciesConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	policies := make([]*models.ServiceAppPolicyConfig, 0)
	for _, policy := range s.cfg.Policies.Policies {
		policies = append(policies, s.policyToRequest(policy))
	}
	req.Policies = policies

	return req
}

func (s sync) syncAppPolicies(ctx context.Context, resource string) error {
	if s.cfg.Policies == nil {
		return nil
	}

	req := s.getAppPoliciesRequest()
	_, err := s.apiClient.CreateAppPoliciesConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
