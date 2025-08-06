package credentials

import (
	"context"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"go.uber.org/zap"
)

// Fetch uses the Azure SDK method NewDefaultAzureCredential to walk the chain of authentication mechanisms to get a credential.
// If running in an Azure VM, it will use the identity assigned to the VM.
// If running locally, it will the identity you have logged into from your local environment.
// For more information, see: https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/authentication-overview
func Fetch(ctx context.Context, logger *zap.Logger) (*azidentity.DefaultAzureCredential, error) {
	// TODO(ja): use pkg/ctx instead
	// l, err := pkgctx.Logger(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	azlog.SetListener(func(event azlog.Event, msg string) {
		logger.Info(msg)
	})
	azlog.SetEvents(azidentity.EventAuthentication)

	return azidentity.NewDefaultAzureCredential(nil)
}
