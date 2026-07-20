package credentials

import (
	"context"
)

func FetchEnv(ctx context.Context, cfg *Config) (map[string]string, error) {
	env := map[string]string{}
	if cfg.ServicePrincipal != nil {
		env["ARM_SUBSCRIPTION_ID"] = cfg.ServicePrincipal.SubscriptionID
		env["ARM_TENANT_ID"] = cfg.ServicePrincipal.SubscriptionTenantID
	}

	// ARM_USE_MSI + ARM_CLIENT_ID makes the provider use this user-assigned identity.
	if cfg.ManagedIdentityClientID != "" {
		env["ARM_USE_MSI"] = "true"
		env["ARM_CLIENT_ID"] = cfg.ManagedIdentityClientID
	}

	return env, nil
}
