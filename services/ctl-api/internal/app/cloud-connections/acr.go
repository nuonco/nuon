package cloudconnections

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

const ACRAudience = "api://AzureADTokenExchange"

var azureUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var azureRegistryIDPattern = regexp.MustCompile(`(?i)^/subscriptions/([0-9a-f-]+)/resourceGroups/([a-z0-9_.()-]{1,90})/providers/Microsoft\.ContainerRegistry/registries/([a-z0-9]{5,50})$`)

func ValidateACRConnection(connection *app.CloudConnection) error {
	if connection.Platform != app.CloudPlatformAzure || connection.Preset != app.CloudConnectionPresetACR {
		return fmt.Errorf("only aws cloud connections support stacks, ecr, and custom; Azure requires the acr preset")
	}
	if !azureUUIDPattern.MatchString(connection.TargetID) {
		return fmt.Errorf("target_id must be an Azure subscription UUID")
	}
	if !azureUUIDPattern.MatchString(connection.Principal) {
		return fmt.Errorf("principal must be the client ID UUID of an existing Microsoft Entra application with an existing service principal")
	}
	if !azureUUIDPattern.MatchString(connection.AzureTenantID) {
		return fmt.Errorf("azure_tenant_id must be a Microsoft Entra tenant UUID")
	}
	parts := azureRegistryIDPattern.FindStringSubmatch(connection.AzureRegistryID)
	if parts == nil || !strings.EqualFold(parts[1], connection.TargetID) || strings.HasSuffix(parts[2], ".") {
		return fmt.Errorf("azure_registry_id must be an exact Azure Container Registry ARM ID in subscription %s", connection.TargetID)
	}
	if err := acr.ValidateLoginServer(connection.AzureRegistryLoginServer); err != nil {
		return err
	}
	if connection.AzureRegistryLoginServer != strings.ToLower(connection.AzureRegistryLoginServer) {
		return fmt.Errorf("azure_registry_login_server must be a lowercase bare <registry>.azurecr.io hostname")
	}
	if connection.GCPProjectNumber != "" {
		return fmt.Errorf("gcp_project_number cannot be used with an Azure ACR connection")
	}
	return nil
}

func ACRConnection(ctx context.Context, db *gorm.DB, orgID string, cfg *app.AzureACRImageConfig, imageURL string) (*app.CloudConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ACR image configuration is required")
	}
	if cfg.TenantID != "" || cfg.ClientID != "" || cfg.ClientSecretName != "" || cfg.ClientCertificateName != "" {
		return nil, fmt.Errorf("ACR cloud connection access cannot be combined with component-level credentials")
	}
	if cfg.ID != "" && cfg.CloudConnectionID == "" {
		return nil, fmt.Errorf("ACR image configuration has no pinned cloud connection; sync the app configuration before building")
	}
	host, repository, ok := strings.Cut(imageURL, "/")
	if !ok || cfg.RegistryURL != host {
		return nil, fmt.Errorf("ACR image URL must match registry_url")
	}
	connection, err := imageConnection(ctx, db, app.CloudConnection{
		ID: cfg.CloudConnectionID, OrgID: orgID, Platform: app.CloudPlatformAzure,
		Preset: app.CloudConnectionPresetACR, AzureRegistryLoginServer: host,
	}, repository)
	if err != nil {
		return nil, err
	}
	if err := ValidateACRConnection(connection); err != nil {
		return nil, fmt.Errorf("invalid ACR cloud connection: %w", err)
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
