package credentials

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// Fetch is used to get credentials, regardless of whether they are in the context, or not.
func Fetch(_ context.Context, cfg *Config) (*azidentity.ClientSecretCredential, error) {
	if cfg.ServicePrincipal == nil || cfg.UseDefault {
		return nil, fmt.Errorf("use default is not supported until the runner has the correct ACR permissions")
	}

	return getServicePrincipalCredentials(cfg.ServicePrincipal)
}

func getServicePrincipalCredentials(creds *ServicePrincipalCredentials) (*azidentity.ClientSecretCredential, error) {
	credential, err := azidentity.NewClientSecretCredential(
		creds.SubscriptionTenantID,
		creds.ServicePrincipalAppID,
		creds.ServicePrincipalPassword,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get credential: %w", err)
	}

	return credential, nil
}
