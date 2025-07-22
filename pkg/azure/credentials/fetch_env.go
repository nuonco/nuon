package credentials

import (
	"context"
)

// NOTE(jm): for now, we only support static service principal credentials
func FetchEnv(ctx context.Context, cfg *Config) (map[string]string, error) {
	if cfg.ServicePrincipal != nil {
		return map[string]string{
			"ARM_SUBSCRIPTION_ID": cfg.ServicePrincipal.SubscriptionID,
			"ARM_TENANT_ID":       cfg.ServicePrincipal.SubscriptionTenantID,
		}, nil
	}
	return map[string]string{}, nil
}
