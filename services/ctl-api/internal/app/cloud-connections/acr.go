package cloudconnections

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

const ACRAudience = "api://AzureADTokenExchange"

func acrHasLegacyAuth(cfg *app.AzureACRImageConfig) bool {
	return cfg.TenantID != "" || cfg.ClientID != "" || cfg.ClientSecretName != "" || cfg.ClientCertificateName != ""
}

func ACRConnection(ctx context.Context, db *gorm.DB, orgID string, cfg *app.AzureACRImageConfig, imageURL string, legacy bool) (*app.CloudConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ACR image configuration is required")
	}
	if acrHasLegacyAuth(cfg) {
		if cfg.CloudConnectionID != "" {
			return nil, stderr.NewInvalidRequest(errors.New("ACR cloud connection access cannot be combined with component-level credentials"))
		}
		return legacyImageAuth(legacy, ErrLegacyImageAuthDisabled)
	}
	if cfg.ID != "" && cfg.CloudConnectionID == "" {
		return legacyImageAuth(legacy, ErrLegacyImageAuthDisabled)
	}
	host, repository, ok := strings.Cut(imageURL, "/")
	if !ok || !strings.EqualFold(cfg.RegistryURL, host) {
		return legacyImageAuth(legacy, stderr.NewInvalidRequest(errors.New("ACR image URL must match registry_url")))
	}
	host = strings.ToLower(host)
	connection, err := imageConnection(ctx, db, app.CloudConnection{
		ID: cfg.CloudConnectionID, OrgID: orgID, Platform: app.CloudPlatformAzure,
		Preset: app.CloudConnectionPresetACR, AzureRegistryLoginServer: host,
	}, repository)
	if errors.Is(err, errNoImageRepositoryMapping) {
		return legacyImageAuth(legacy, err)
	}
	if err != nil {
		return nil, err
	}
	cfg.CloudConnectionID = connection.ID
	return connection, nil
}

func ACRCredential(issuer *oidcissuer.Issuer, connection *app.CloudConnection, subject string) (*azidentity.ClientAssertionCredential, error) {
	if issuer == nil {
		return nil, fmt.Errorf("cloud connection OIDC issuer is unavailable")
	}
	return azidentity.NewClientAssertionCredential(connection.AzureTenantID, connection.Principal, func(ctx context.Context) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return issuer.Mint(ctx, subject, ACRAudience, 10*time.Minute)
	}, &azidentity.ClientAssertionCredentialOptions{ClientOptions: azcore.ClientOptions{
		Transport: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}})
}
