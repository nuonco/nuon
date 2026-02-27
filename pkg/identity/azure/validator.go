package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/identity"
	"github.com/nuonco/nuon/pkg/identity/jwks"
)

// Validator validates Azure managed identity tokens
type Validator struct {
	logger    *zap.Logger
	jwksCache *jwks.Cache
	audience  string
}

// customClaims represents Azure-specific JWT claims
type customClaims struct {
	TenantID             string `json:"tid"`
	ObjectID             string `json:"oid"`
	ManagedIdentityResID string `json:"xms_mirid"`
	AzureResourceID      string `json:"xms_az_rid"`
}

// Validate implements validator.CustomClaims interface
func (c *customClaims) Validate(ctx context.Context) error {
	return nil
}

// NewValidator creates a new Azure identity validator
func NewValidator(logger *zap.Logger, jwksCache *jwks.Cache, audience string) *Validator {
	return &Validator{
		logger:    logger,
		jwksCache: jwksCache,
		audience:  audience,
	}
}

// ValidateToken validates an Azure managed identity token
func (v *Validator) ValidateToken(ctx context.Context, token string) (*identity.Claims, error) {
	// Parse token to get tenant ID
	unverified, err := parseUnverifiedJWT(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract tenant ID from issuer
	tenantID, err := extractTenantFromIssuer(unverified.Issuer)
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
			return &customClaims{}
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
	customClaims := claims.CustomClaims.(*customClaims)

	// Extract subscription ID from managed identity resource ID
	var subscriptionID string
	if customClaims.ManagedIdentityResID != "" {
		subscriptionID = extractSubscriptionFromResourceID(customClaims.ManagedIdentityResID)
	}

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"tid":             customClaims.TenantID,
		"oid":             customClaims.ObjectID,
		"xms_mirid":       customClaims.ManagedIdentityResID,
		"xms_az_rid":      customClaims.AzureResourceID,
		"subscription_id": subscriptionID,
	}

	return &identity.Claims{
		Subject:   claims.RegisteredClaims.Subject,
		Issuer:    claims.RegisteredClaims.Issuer,
		Audience:  claims.RegisteredClaims.Audience,
		ExpiresAt: time.Unix(claims.RegisteredClaims.Expiry, 0),
		Provider:  identity.ProviderTypeAzure,
		Custom:    customClaimsMap,
	}, nil
}

// ValidateIdentity validates that Azure token claims match the expected runner identity
func (v *Validator) ValidateIdentity(ctx context.Context, claims *identity.Claims, valCtx *identity.ValidationContext) error {
	tenantID, ok := claims.Custom["tid"].(string)
	if !ok || tenantID == "" {
		return fmt.Errorf("tenant ID not found in claims")
	}

	subscriptionID, _ := claims.Custom["subscription_id"].(string)

	v.logger.Debug("azure identity validated",
		zap.String("tenant_id", tenantID),
		zap.String("subscription_id", subscriptionID),
		zap.String("runner_id", valCtx.RunnerID))

	return nil
}

// Provider returns the provider type
func (v *Validator) Provider() identity.ProviderType {
	return identity.ProviderTypeAzure
}

// extractSubscriptionFromResourceID extracts subscription ID from Azure resource ID
// Format: /subscriptions/{subscription-id}/resourceGroups/{resource-group}/...
func extractSubscriptionFromResourceID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if part == "subscriptions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractTenantFromIssuer extracts the tenant ID from an Azure AD issuer URL
// Example: https://login.microsoftonline.com/{tenant}/v2.0 -> {tenant}
func extractTenantFromIssuer(issuer string) (string, error) {
	// Remove https:// prefix
	path := strings.TrimPrefix(issuer, "https://")
	path = strings.TrimPrefix(path, "http://")

	// Remove domain
	path = strings.TrimPrefix(path, "login.microsoftonline.com/")
	path = strings.TrimPrefix(path, "sts.windows.net/")

	// Remove /v2.0 or /v1.0 suffix
	path = strings.TrimSuffix(path, "/v2.0")
	path = strings.TrimSuffix(path, "/v1.0")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return "", fmt.Errorf("unable to extract tenant ID from Azure issuer: %s", issuer)
	}

	return path, nil
}
