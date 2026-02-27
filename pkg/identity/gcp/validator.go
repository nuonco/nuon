package gcp

import (
	"context"
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/identity"
	"github.com/nuonco/nuon/pkg/identity/jwks"
)

const (
	// Google's OIDC issuer
	googleIssuer = "https://accounts.google.com"
)

// Validator validates GCP instance identity tokens
type Validator struct {
	logger    *zap.Logger
	jwksCache *jwks.Cache
	audience  string
}

// customClaims represents GCP-specific JWT claims
type customClaims struct {
	Email           string                 `json:"email"`
	EmailVerified   bool                   `json:"email_verified"`
	Google          map[string]interface{} `json:"google"`
	AuthorizedParty string                 `json:"azp"`
}

// Validate implements validator.CustomClaims interface
func (c *customClaims) Validate(ctx context.Context) error {
	return nil
}

// NewValidator creates a new GCP identity validator
func NewValidator(logger *zap.Logger, jwksCache *jwks.Cache, audience string) *Validator {
	return &Validator{
		logger:    logger,
		jwksCache: jwksCache,
		audience:  audience,
	}
}

// ValidateToken validates a GCP instance identity token
func (v *Validator) ValidateToken(ctx context.Context, token string) (*identity.Claims, error) {
	provider, err := v.jwksCache.GetProvider(ctx, googleIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS provider: %w", err)
	}

	// Create JWT validator with custom claims
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		googleIssuer,
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

	// Extract project ID from Google compute engine claims
	var projectID string
	if googleClaims, ok := customClaims.Google["compute_engine"].(map[string]interface{}); ok {
		if pid, ok := googleClaims["project_id"].(string); ok {
			projectID = pid
		}
	}

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"email":          customClaims.Email,
		"email_verified": customClaims.EmailVerified,
		"google":         customClaims.Google,
		"project_id":     projectID,
	}

	return &identity.Claims{
		Subject:   claims.RegisteredClaims.Subject,
		Issuer:    claims.RegisteredClaims.Issuer,
		Audience:  claims.RegisteredClaims.Audience,
		ExpiresAt: time.Unix(claims.RegisteredClaims.Expiry, 0),
		Provider:  identity.ProviderTypeGCP,
		Custom:    customClaimsMap,
	}, nil
}

// ValidateIdentity validates that GCP token claims match the expected runner identity
func (v *Validator) ValidateIdentity(ctx context.Context, claims *identity.Claims, valCtx *identity.ValidationContext) error {
	// Extract project ID
	projectID, ok := claims.Custom["project_id"].(string)
	if !ok || projectID == "" {
		return fmt.Errorf("project ID not found in token claims")
	}

	// Extract service account email
	email, ok := claims.Custom["email"].(string)
	if !ok || email == "" {
		return fmt.Errorf("service account email not found in token claims")
	}

	// Verify email is verified
	emailVerified, ok := claims.Custom["email_verified"].(bool)
	if !ok || !emailVerified {
		return fmt.Errorf("service account email not verified")
	}

	v.logger.Debug("gcp identity validated",
		zap.String("project_id", projectID),
		zap.String("service_account", email),
		zap.String("runner_id", valCtx.RunnerID))

	return nil
}

// Provider returns the provider type
func (v *Validator) Provider() identity.ProviderType {
	return identity.ProviderTypeGCP
}
