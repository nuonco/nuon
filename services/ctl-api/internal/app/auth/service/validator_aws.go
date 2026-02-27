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

// AWSOIDCValidator validates AWS EC2 instance identity tokens
type AWSOIDCValidator struct {
	l         *zap.Logger
	db        *gorm.DB
	jwksCache *JWKSCache
}

// awsCustomClaims represents AWS-specific JWT claims
type awsCustomClaims struct {
	AccountID  string `json:"accountId"`
	InstanceID string `json:"instanceId"`
	RoleARN    string `json:"roleArn"`
}

// Validate implements validator.CustomClaims interface
func (c *awsCustomClaims) Validate(ctx context.Context) error {
	return nil
}

// NewAWSOIDCValidator creates a new AWS OIDC validator
func NewAWSOIDCValidator(l *zap.Logger, db *gorm.DB, jwksCache *JWKSCache) *AWSOIDCValidator {
	return &AWSOIDCValidator{
		l:         l,
		db:        db,
		jwksCache: jwksCache,
	}
}

// ValidateToken validates an AWS EC2 instance identity token
func (v *AWSOIDCValidator) ValidateToken(ctx context.Context, token string) (*OIDCClaims, error) {
	// Parse token without validation to extract region from issuer
	unverified, err := parseUnverifiedJWT(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract region from issuer URL
	region, err := extractRegionFromAWSIssuer(unverified.Issuer)
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
	// Note: AWS EC2 instance identity documents may not have a strict audience requirement
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{}, // AWS doesn't always use audience for EC2 instance identity
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &awsCustomClaims{}
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
	customClaims := claims.CustomClaims.(*awsCustomClaims)

	// Build custom claims map
	customClaimsMap := map[string]interface{}{
		"accountId":  customClaims.AccountID,
		"instanceId": customClaims.InstanceID,
		"roleArn":    customClaims.RoleARN,
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

// ValidateRunnerIdentity validates that AWS token claims match the runner's install configuration
func (v *AWSOIDCValidator) ValidateRunnerIdentity(ctx context.Context, runner *app.Runner, claims *OIDCClaims) error {
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

	if installStack.InstallStackOutputs.AWSStackOutputs == nil {
		return fmt.Errorf("install does not have AWS stack outputs")
	}

	awsOutputs := installStack.InstallStackOutputs.AWSStackOutputs

	// Validate account ID
	accountID, ok := claims.Custom["accountId"].(string)
	if !ok {
		return fmt.Errorf("account ID not found in token claims")
	}

	if accountID != awsOutputs.AccountID {
		return fmt.Errorf("account ID mismatch: got %s, expected %s", accountID, awsOutputs.AccountID)
	}

	// Validate instance ID format
	instanceID := claims.Subject
	if instanceID == "" {
		// Try to get from custom claims
		instanceID, _ = claims.Custom["instanceId"].(string)
	}

	if instanceID != "" && !ec2InstanceIDPattern.MatchString(instanceID) {
		return fmt.Errorf("invalid instance ID format: %s", instanceID)
	}

	// Validate IAM role if present in claims
	if roleARN, ok := claims.Custom["roleArn"].(string); ok && roleARN != "" {
		expectedRole := awsOutputs.RunnerIAMRoleARN
		if expectedRole != "" {
			roleName, err := extractRoleNameFromARN(roleARN)
			if err != nil {
				return fmt.Errorf("failed to parse role ARN: %w", err)
			}
			expectedRoleName, err := extractRoleNameFromARN(expectedRole)
			if err != nil {
				return fmt.Errorf("failed to parse expected role ARN: %w", err)
			}
			if roleName != expectedRoleName {
				return fmt.Errorf("IAM role mismatch: got %s, expected %s", roleName, expectedRoleName)
			}
		}
	}

	// Cross-validate region if available in install outputs
	if awsOutputs.Region != "" {
		// Extract region from issuer
		issuerRegion, err := extractRegionFromAWSIssuer(claims.Issuer)
		if err == nil && issuerRegion != awsOutputs.Region {
			v.l.Warn("region mismatch between token and install",
				zap.String("token_region", issuerRegion),
				zap.String("install_region", awsOutputs.Region))
			// This is a warning, not an error, as tokens may be issued from different regions
		}
	}

	return nil
}

// extractRoleNameFromARN extracts the role name from an IAM role ARN
// Handles both assumed-role and role ARNs
func extractRoleNameFromARN(arn string) (string, error) {
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

// Helper methods (reuse same pattern)

func (v *AWSOIDCValidator) getInstallByRunnerGroup(ctx context.Context, runnerGroup *app.RunnerGroup) (*app.Install, error) {
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

func (v *AWSOIDCValidator) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
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
