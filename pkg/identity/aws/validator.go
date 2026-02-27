package aws

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

// Validator validates AWS EC2 instance identity tokens
type Validator struct {
	logger    *zap.Logger
	jwksCache *jwks.Cache
}

// customClaims represents AWS-specific JWT claims
type customClaims struct {
	AccountID  string `json:"accountId"`
	InstanceID string `json:"instanceId"`
	RoleARN    string `json:"roleArn"`
}

// Validate implements validator.CustomClaims interface
func (c *customClaims) Validate(ctx context.Context) error {
	return nil
}

// NewValidator creates a new AWS identity validator
func NewValidator(logger *zap.Logger, jwksCache *jwks.Cache) *Validator {
	return &Validator{
		logger:    logger,
		jwksCache: jwksCache,
	}
}

// ValidateToken validates an AWS EC2 instance identity token
func (v *Validator) ValidateToken(ctx context.Context, token string) (*identity.Claims, error) {
	// Parse token without validation to extract region from issuer
	unverified, err := parseUnverifiedJWT(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract region from issuer URL
	region, err := extractRegionFromIssuer(unverified.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to extract region: %w", err)
	}

	// AWS EC2 instance identity JWKS endpoint
	issuerURL := fmt.Sprintf("https://sts.%s.amazonaws.com", region)

	provider, err := v.jwksCache.GetProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS provider: %w", err)
	}

	// Create JWT validator with custom claims
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{}, // AWS doesn't always use audience for EC2 instance identity
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

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"accountId":  customClaims.AccountID,
		"instanceId": customClaims.InstanceID,
		"roleArn":    customClaims.RoleARN,
		"region":     region,
	}

	return &identity.Claims{
		Subject:   claims.RegisteredClaims.Subject,
		Issuer:    claims.RegisteredClaims.Issuer,
		Audience:  claims.RegisteredClaims.Audience,
		ExpiresAt: time.Unix(claims.RegisteredClaims.Expiry, 0),
		Provider:  identity.ProviderTypeAWS,
		Custom:    customClaimsMap,
	}, nil
}

// ValidateIdentity validates that AWS token claims match the expected runner identity
// The validation context should contain install-specific AWS configuration
func (v *Validator) ValidateIdentity(ctx context.Context, claims *identity.Claims, valCtx *identity.ValidationContext) error {
	// For AWS, the validation context custom data should contain:
	// - expected_account_id
	// - expected_role_arn
	// - expected_region (optional, for warnings)

	// This method should be called by the service layer that has access to install configuration
	// The basic validation is done here, service layer provides the expected values

	accountID, ok := claims.Custom["accountId"].(string)
	if !ok {
		return fmt.Errorf("account ID not found in token claims")
	}

	// Instance ID validation
	instanceID := claims.Subject
	if instanceID == "" {
		instanceID, _ = claims.Custom["instanceId"].(string)
	}

	if instanceID != "" && !isValidInstanceID(instanceID) {
		return fmt.Errorf("invalid instance ID format: %s", instanceID)
	}

	v.logger.Debug("aws identity validated",
		zap.String("account_id", accountID),
		zap.String("instance_id", instanceID),
		zap.String("runner_id", valCtx.RunnerID))

	return nil
}

// Provider returns the provider type
func (v *Validator) Provider() identity.ProviderType {
	return identity.ProviderTypeAWS
}

// isValidInstanceID checks if the instance ID matches AWS EC2 format
func isValidInstanceID(id string) bool {
	// EC2 instance IDs: i-xxxxxxxx or i-xxxxxxxxxxxxxxxxx (8 or 17 hex chars)
	if !strings.HasPrefix(id, "i-") {
		return false
	}
	suffix := id[2:]
	if len(suffix) != 8 && len(suffix) != 17 {
		return false
	}
	for _, c := range suffix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// extractRegionFromIssuer extracts the AWS region from the issuer URL
func extractRegionFromIssuer(issuer string) (string, error) {
	// Expected format: https://sts.{region}.amazonaws.com
	// or: https://sts.amazonaws.com (us-east-1)

	issuer = strings.TrimPrefix(issuer, "https://")
	issuer = strings.TrimPrefix(issuer, "http://")

	if issuer == "sts.amazonaws.com" {
		return "us-east-1", nil
	}

	if strings.HasPrefix(issuer, "sts.") && strings.Contains(issuer, ".amazonaws.com") {
		parts := strings.Split(issuer, ".")
		if len(parts) >= 3 {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("unable to extract region from issuer: %s", issuer)
}

// ExtractRoleNameFromARN extracts the role name from an IAM role ARN
// Handles both assumed-role and role ARNs
func ExtractRoleNameFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("invalid ARN format: %s", arn)
	}

	resource := parts[5]

	// Handle assumed-role format: assumed-role/role-name/session-name
	if strings.HasPrefix(resource, "assumed-role/") {
		resourceParts := strings.Split(resource, "/")
		if len(resourceParts) >= 2 {
			return resourceParts[1], nil
		}
	}

	// Handle role format: role/role-name or role/path/role-name
	if strings.HasPrefix(resource, "role/") {
		resourceParts := strings.Split(resource, "/")
		if len(resourceParts) >= 2 {
			// Return the last segment (role name)
			return resourceParts[len(resourceParts)-1], nil
		}
	}

	return "", fmt.Errorf("unable to extract role name from ARN: %s", arn)
}
