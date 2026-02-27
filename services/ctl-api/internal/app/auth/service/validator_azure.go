package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AzureOIDCValidator validates Azure managed identity tokens
type AzureOIDCValidator struct {
	l         *zap.Logger
	db        *gorm.DB
	jwksCache *JWKSCache
	audience  string
}

// azureCustomClaims represents Azure-specific JWT claims
type azureCustomClaims struct {
	TenantID             string `json:"tid"`
	ObjectID             string `json:"oid"`
	ManagedIdentityResID string `json:"xms_mirid"`
	AzureResourceID      string `json:"xms_az_rid"`
}

// Validate implements validator.CustomClaims interface
func (c *azureCustomClaims) Validate(ctx context.Context) error {
	return nil
}

// NewAzureOIDCValidator creates a new Azure OIDC validator
func NewAzureOIDCValidator(l *zap.Logger, db *gorm.DB, jwksCache *JWKSCache, audience string) *AzureOIDCValidator {
	return &AzureOIDCValidator{
		l:         l,
		db:        db,
		jwksCache: jwksCache,
		audience:  audience,
	}
}

// ValidateToken validates an Azure managed identity token
func (v *AzureOIDCValidator) ValidateToken(ctx context.Context, token string) (*OIDCClaims, error) {
	// Parse token to get tenant ID
	unverified, err := parseUnverifiedJWT(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract tenant ID from issuer
	tenantID, err := extractTenantFromAzureIssuer(unverified.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tenant ID: %w", err)
	}

	// Azure AD OIDC discovery endpoint
	issuerURL := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID)

	provider, err := v.jwksCache.GetProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS provider: %w", err)
	}

	// Create JWT validator with custom claims
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{v.audience},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &azureCustomClaims{}
		}),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validator: %w", err)
	}

	validatedToken, err := jwtValidator.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims := validatedToken.(*validator.ValidatedClaims)
	customClaims := claims.CustomClaims.(*azureCustomClaims)

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"tid":        customClaims.TenantID,
		"oid":        customClaims.ObjectID,
		"xms_mirid":  customClaims.ManagedIdentityResID,
		"xms_az_rid": customClaims.AzureResourceID,
	}

	return &OIDCClaims{
		Subject:   claims.RegisteredClaims.Subject,
		Issuer:    claims.RegisteredClaims.Issuer,
		Audience:  claims.RegisteredClaims.Audience,
		ExpiresAt: claims.RegisteredClaims.Expiry,
		IssuedAt:  claims.RegisteredClaims.IssuedAt,
		Custom:    customClaimsMap,
	}, nil
}

// ValidateRunnerIdentity validates that Azure token claims match the runner's install configuration
func (v *AzureOIDCValidator) ValidateRunnerIdentity(ctx context.Context, runner *app.Runner, claims *OIDCClaims) error {
	// Get install for runner
	install, err := v.getInstallByRunnerGroup(ctx, &runner.RunnerGroup)
	if err != nil {
		return fmt.Errorf("failed to get install: %w", err)
	}

	// Get install stack outputs
	installStack, err := v.getInstallStackWithOutputs(ctx, install.ID)
	if err != nil {
		return fmt.Errorf("failed to get install stack: %w", err)
	}

	if installStack.InstallStackOutputs.AzureStackOutputs == nil {
		return fmt.Errorf("install does not have Azure stack outputs")
	}

	azureOutputs := installStack.InstallStackOutputs.AzureStackOutputs

	// Validate tenant ID
	tenantID, ok := claims.Custom["tid"].(string)
	if !ok {
		return fmt.Errorf("tenant ID not found in claims")
	}

	if tenantID != azureOutputs.SubscriptionTenantID {
		return fmt.Errorf("tenant ID mismatch: got %s, expected %s", tenantID, azureOutputs.SubscriptionTenantID)
	}

	// Validate subscription ID if managed identity resource ID is present
	if mirid, ok := claims.Custom["xms_mirid"].(string); ok && mirid != "" {
		// Format: /subscriptions/{subscription-id}/resourceGroups/{resource-group}/...
		if !strings.Contains(mirid, azureOutputs.SubscriptionID) {
			return fmt.Errorf("subscription ID mismatch in resource ID")
		}
	}

	return nil
}

// Helper methods (reuse same pattern as GCP validator)

func (v *AzureOIDCValidator) getInstallByRunnerGroup(ctx context.Context, runnerGroup *app.RunnerGroup) (*app.Install, error) {
	if runnerGroup.OwnerType != "installs" {
		return nil, fmt.Errorf("runner group is not associated with an install")
	}

	var install app.Install
	res := v.db.WithContext(ctx).First(&install, "id = ?", runnerGroup.OwnerID)
	if res.Error != nil {
		return nil, res.Error
	}
	return &install, nil
}

func (v *AzureOIDCValidator) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
	var version app.InstallStackVersion
	res := v.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Where("status->>'status' = ?", app.InstallStackVersionStatusActive).
		Order("created_at DESC").
		First(&version)
	if res.Error != nil {
		return nil, fmt.Errorf("no active install stack version found: %w", res.Error)
	}

	var installStack app.InstallStack
	res = v.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		First(&installStack, "id = ?", version.InstallStackID)
	if res.Error != nil {
		return nil, res.Error
	}

	return &installStack, nil
}
