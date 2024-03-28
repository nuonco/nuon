package provision

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type CheckAzurePrincipalRequest struct {
	Location                 string `validate:"required"`
	SubscriptionID           string `validate:"required"`
	SubscriptionTenantID     string `validate:"required"`
	ServicePrincipalAppID    string `validate:"required"`
	ServicePrincipalPassword string `validate:"required"`
}

type CheckAzurePrincipalResponse struct{}

func (a *Activities) CheckAzurePrincipal(ctx context.Context, req CheckAzurePrincipalRequest) (*CheckAzurePrincipalResponse, error) {
	credential, err := azidentity.NewClientSecretCredential(
		req.SubscriptionTenantID,
		req.ServicePrincipalAppID,
		req.ServicePrincipalPassword,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get credential: %w", err)
	}

	_, err = credential.GetToken(ctx,
		policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get token from credential: %w", err)
	}

	return nil, nil
}
