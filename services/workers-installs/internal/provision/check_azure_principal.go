package provision

import (
	"context"
	"fmt"
	"os"

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
	// NOTE(jm): the azure sdk does not have a way to directly set these environment variables, therefore we need to
	// manually set them, and then unset them. Ultimately, this is not going to be thread/process safe, but for now
	// it's the best we can do.

	os.Setenv("ARM_SUBSCRIPTION_ID", req.SubscriptionID)
	os.Setenv("ARM_TENANT_ID", req.SubscriptionTenantID)
	os.Setenv("ARM_CLIENT_ID", req.ServicePrincipalAppID)
	os.Setenv("ARM_CLIENT_SECRET", req.ServicePrincipalPassword)

	_, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get credential: %w", err)
	}

	os.Unsetenv("ARM_SUBSCRIPTION_ID")
	os.Unsetenv("ARM_TENANT_ID")
	os.Unsetenv("ARM_CLIENT_ID")
	os.Unsetenv("ARM_CLIENT_SECRET")

	return nil, nil
}
