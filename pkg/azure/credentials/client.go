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

func Fetch(ctx context.Context, cfg *Config, logger *zap.Logger) (azcore.TokenCredential, error) {
	azlog.SetListener(func(event azlog.Event, msg string) {
		logger.Info(msg)
	})
	azlog.SetEvents(azidentity.EventAuthentication)

	if cfg != nil && cfg.HasAppRegistrationCredentials() {
		return fetchAppRegistration(cfg, logger)
	}

	// why: In local dev, skip ManagedIdentityCredential to avoid a ~30s IMDS
	// timeout that exhausts the job context before AzureCLICredential runs.
	if os.Getenv("ENV") == "development" {
		logger.Info("local dev: using AzureCLICredential (skipping ManagedIdentity)")
		return azidentity.NewAzureCLICredential(nil)
	}

	return azidentity.NewDefaultAzureCredential(nil)
}

func fetchAppRegistration(cfg *Config, logger *zap.Logger) (azcore.TokenCredential, error) {
	if len(cfg.ClientCertificatePEM) > 0 {
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
