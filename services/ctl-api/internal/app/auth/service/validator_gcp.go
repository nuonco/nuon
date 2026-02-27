package service

import (
	"context"
	"fmt"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GCPOIDCValidator validates GCP instance identity tokens
type GCPOIDCValidator struct {
	l         *zap.Logger
	db        *gorm.DB
	jwksCache *JWKSCache
	audience  string
}

// gcpCustomClaims represents GCP-specific JWT claims
type gcpCustomClaims struct {
	Email           string                 `json:"email"`
	EmailVerified   bool                   `json:"email_verified"`
	Google          map[string]interface{} `json:"google"`
	AuthorizedParty string                 `json:"azp"`
	IssuedTo        string                 `json:"aud"`
}

// Validate implements validator.CustomClaims interface
func (c *gcpCustomClaims) Validate(ctx context.Context) error {
	return nil
}

// NewGCPOIDCValidator creates a new GCP OIDC validator
func NewGCPOIDCValidator(l *zap.Logger, db *gorm.DB, jwksCache *JWKSCache, audience string) *GCPOIDCValidator {
	return &GCPOIDCValidator{
		l:         l,
		db:        db,
		jwksCache: jwksCache,
		audience:  audience,
	}
}

// ValidateToken validates a GCP instance identity token
func (v *GCPOIDCValidator) ValidateToken(ctx context.Context, token string) (*OIDCClaims, error) {
	// GCP instance identity tokens are signed by Google
	issuerURL := "https://accounts.google.com"

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
			return &gcpCustomClaims{}
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
	customClaims := claims.CustomClaims.(*gcpCustomClaims)

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"email":          customClaims.Email,
		"email_verified": customClaims.EmailVerified,
		"google":         customClaims.Google,
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

// ValidateRunnerIdentity validates that GCP token claims match the runner's install configuration
func (v *GCPOIDCValidator) ValidateRunnerIdentity(ctx context.Context, runner *app.Runner, claims *OIDCClaims) error {
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

	if installStack.InstallStackOutputs.GCPStackOutputs == nil {
		return fmt.Errorf("install does not have GCP stack outputs")
	}

	gcpOutputs := installStack.InstallStackOutputs.GCPStackOutputs

	// Extract project ID from Google compute engine claims
	googleClaims, ok := claims.Custom["google"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing google compute engine claims")
	}

	computeEngine, ok := googleClaims["compute_engine"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing compute_engine claims")
	}

	projectID, ok := computeEngine["project_id"].(string)
	if !ok {
		return fmt.Errorf("project ID not found in token claims")
	}

	// Validate project ID
	if projectID != gcpOutputs.ProjectID {
		return fmt.Errorf("project ID mismatch: got %s, expected %s", projectID, gcpOutputs.ProjectID)
	}

	// Validate service account email if configured
	if gcpOutputs.RunnerServiceAccount != "" {
		email, ok := claims.Custom["email"].(string)
		if !ok {
			return fmt.Errorf("email not found in token claims")
		}
		if email != gcpOutputs.RunnerServiceAccount {
			return fmt.Errorf("service account mismatch: got %s, expected %s", email, gcpOutputs.RunnerServiceAccount)
		}
	}

	return nil
}

// Helper methods

func (v *GCPOIDCValidator) getInstallByRunnerGroup(ctx context.Context, runnerGroup *app.RunnerGroup) (*app.Install, error) {
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

func (v *GCPOIDCValidator) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
	// Get the most recent active install stack version
	var version app.InstallStackVersion
	res := v.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Where("status->>'status' = ?", app.InstallStackVersionStatusActive).
		Order("created_at DESC").
		First(&version)
	if res.Error != nil {
		return nil, fmt.Errorf("no active install stack version found: %w", res.Error)
	}

	// Get install stack with outputs
	var installStack app.InstallStack
	res = v.db.WithContext(ctx).
		Preload("InstallStackOutputs").
		First(&installStack, "id = ?", version.InstallStackID)
	if res.Error != nil {
		return nil, res.Error
	}

	return &installStack, nil
}
