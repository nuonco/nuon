package credentials

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"go.uber.org/zap"
)

// Fetch uses the Azure SDK method NewDefaultAzureCredential to walk the chain of authentication mechanisms to get a credential.
// If running in an Azure VM, it will use the identity assigned to the VM.
// If running locally, it will the identity you have logged into from your local environment.
// For more information, see: https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/authentication-overview
func Fetch(ctx context.Context, logger *zap.Logger) (azcore.TokenCredential, error) {
	azlog.SetListener(func(event azlog.Event, msg string) {
		logger.Info(msg)
	})
	azlog.SetEvents(azidentity.EventAuthentication)

	// In local dev, skip ManagedIdentityCredential to avoid a ~30s IMDS
	// timeout that exhausts the job context before AzureCLICredential runs.
	if os.Getenv("ENV") == "development" {
		logger.Info("local dev: using AzureCLICredential (skipping ManagedIdentity)")
		return azidentity.NewAzureCLICredential(nil)
	}

	return azidentity.NewDefaultAzureCredential(nil)
}
