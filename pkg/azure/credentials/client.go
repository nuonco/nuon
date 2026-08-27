package credentials

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"go.uber.org/zap"
)

// Fetch returns a credential for cfg.
//
// When cfg names an app registration with a secret or certificate, that is
// used — it is the only way to authenticate into a tenant this process has no
// identity in, which is what reaching a vendor's registry requires.
//
// Otherwise it falls back to the ambient identity via NewDefaultAzureCredential:
// on an Azure VM that is the identity assigned to the VM, and locally it is
// whatever you are logged in as. A nil cfg selects the ambient path, so callers
// with no credential to offer can keep passing nil.
// For more information, see: https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/authentication-overview
func Fetch(ctx context.Context, cfg *Config, logger *zap.Logger) (azcore.TokenCredential, error) {
	azlog.SetListener(func(event azlog.Event, msg string) {
		logger.Info(msg)
	})
	azlog.SetEvents(azidentity.EventAuthentication)

	if cfg != nil && cfg.HasAppRegistrationCredentials() {
		return fetchAppRegistration(cfg, logger)
	}

	// In local dev, skip ManagedIdentityCredential to avoid a ~30s IMDS
	// timeout that exhausts the job context before AzureCLICredential runs.
	if os.Getenv("ENV") == "development" {
		logger.Info("local dev: using AzureCLICredential (skipping ManagedIdentity)")
		return azidentity.NewAzureCLICredential(nil)
	}

	return azidentity.NewDefaultAzureCredential(nil)
}

func fetchAppRegistration(cfg *Config, logger *zap.Logger) (azcore.TokenCredential, error) {
	if len(cfg.ClientCertificatePEM) > 0 {
		// Passphrase-protected keys are not supported: there is nowhere in the
		// config to put a passphrase, so a caller supplying one would fail here
		// with a confusing parse error rather than a missing-field error.
		certs, key, err := azidentity.ParseCertificates(cfg.ClientCertificatePEM, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to parse client certificate for app registration %s: %w", cfg.ClientID, err)
		}

		logger.Info("authenticating as app registration via certificate",
			zap.String("client_id", cfg.ClientID),
			zap.String("tenant_id", cfg.TenantID),
		)
		return azidentity.NewClientCertificateCredential(cfg.TenantID, cfg.ClientID, certs, key, nil)
	}

	logger.Info("authenticating as app registration via client secret",
		zap.String("client_id", cfg.ClientID),
		zap.String("tenant_id", cfg.TenantID),
	)
	return azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
}
