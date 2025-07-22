package credentials

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// Fetch uses the Azure SDK method NewDefaultAzureCredential to walk the chain of authentication mechanisms to get a credential.
// If running in an Azure VM, it will use the identity assigned to the VM.
// If running locally, it will the identity you have logged into from your local environment.
// For more information, see: https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/authentication-overview
func Fetch(_ context.Context) (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)
}
