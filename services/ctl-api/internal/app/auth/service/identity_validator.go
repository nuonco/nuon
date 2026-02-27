package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/identity"
	awsidentity "github.com/nuonco/nuon/pkg/identity/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// IdentityValidator wraps pkg/identity validators with install-specific validation logic
type IdentityValidator struct {
	logger    *zap.Logger
	db        *gorm.DB
	validator identity.Validator
}

// NewIdentityValidator creates a new identity validator with install validation
func NewIdentityValidator(logger *zap.Logger, db *gorm.DB, validator identity.Validator) *IdentityValidator {
	return &IdentityValidator{
		logger:    logger,
		db:        db,
		validator: validator,
	}
}

// ValidateToken validates the token using the underlying validator
func (v *IdentityValidator) ValidateToken(ctx context.Context, token string) (*identity.Claims, error) {
	return v.validator.ValidateToken(ctx, token)
}

// ValidateRunnerIdentity validates that the identity claims match the runner's install configuration
func (v *IdentityValidator) ValidateRunnerIdentity(ctx context.Context, runner *app.Runner, claims *identity.Claims) error {
	// Create validation context
	valCtx := &identity.ValidationContext{
		RunnerID:  runner.ID,
		InstallID: "", // Will be populated from runner group
		OrgID:     runner.RunnerGroup.OrgID,
	}

	// Get install for runner
	install, err := v.getInstallByRunnerGroup(ctx, &runner.RunnerGroup)
	if err != nil {
		return fmt.Errorf("failed to get install: %w", err)
	}
	valCtx.InstallID = install.ID

	// First, do basic validation through the provider
	if err := v.validator.ValidateIdentity(ctx, claims, valCtx); err != nil {
		return err
	}

	// Then do provider-specific install validation
	switch claims.Provider {
	case identity.ProviderTypeAWS:
		return v.validateAWSInstall(ctx, install.ID, claims)
	case identity.ProviderTypeGCP:
		return v.validateGCPInstall(ctx, install.ID, claims)
	case identity.ProviderTypeAzure:
		return v.validateAzureInstall(ctx, install.ID, claims)
	case identity.ProviderTypeLocal:
		// Local provider doesn't need install validation
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", claims.Provider)
	}
}

// validateAWSInstall validates AWS-specific install configuration
func (v *IdentityValidator) validateAWSInstall(ctx context.Context, installID string, claims *identity.Claims) error {
	installStack, err := v.getInstallStackWithOutputs(ctx, installID)
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

	// Validate IAM role if present
	if roleARN, ok := claims.Custom["roleArn"].(string); ok && roleARN != "" && awsOutputs.RunnerIAMRoleARN != "" {
		roleName, err := awsidentity.ExtractRoleNameFromARN(roleARN)
		if err != nil {
			return fmt.Errorf("failed to parse role ARN: %w", err)
		}
		expectedRoleName, err := awsidentity.ExtractRoleNameFromARN(awsOutputs.RunnerIAMRoleARN)
		if err != nil {
			return fmt.Errorf("failed to parse expected role ARN: %w", err)
		}
		if roleName != expectedRoleName {
			return fmt.Errorf("IAM role mismatch: got %s, expected %s", roleName, expectedRoleName)
		}
	}

	// Cross-validate region
	if region, ok := claims.Custom["region"].(string); ok && region != "" && awsOutputs.Region != "" {
		if region != awsOutputs.Region {
			v.logger.Warn("region mismatch between token and install",
				zap.String("token_region", region),
				zap.String("install_region", awsOutputs.Region))
		}
	}

	return nil
}

// validateGCPInstall validates GCP-specific install configuration
func (v *IdentityValidator) validateGCPInstall(ctx context.Context, installID string, claims *identity.Claims) error {
	installStack, err := v.getInstallStackWithOutputs(ctx, installID)
	if err != nil {
		return fmt.Errorf("failed to get install stack: %w", err)
	}

	if installStack.InstallStackOutputs.GCPStackOutputs == nil {
		return fmt.Errorf("install does not have GCP stack outputs")
	}

	gcpOutputs := installStack.InstallStackOutputs.GCPStackOutputs

	// Validate project ID
	projectID, ok := claims.Custom["project_id"].(string)
	if !ok || projectID == "" {
		return fmt.Errorf("project ID not found in token claims")
	}

	if projectID != gcpOutputs.ProjectID {
		return fmt.Errorf("project ID mismatch: got %s, expected %s", projectID, gcpOutputs.ProjectID)
	}

	// Validate service account if configured
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

// validateAzureInstall validates Azure-specific install configuration
func (v *IdentityValidator) validateAzureInstall(ctx context.Context, installID string, claims *identity.Claims) error {
	installStack, err := v.getInstallStackWithOutputs(ctx, installID)
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

	// Validate subscription ID if present
	if subscriptionID, ok := claims.Custom["subscription_id"].(string); ok && subscriptionID != "" {
		if subscriptionID != azureOutputs.SubscriptionID {
			return fmt.Errorf("subscription ID mismatch: got %s, expected %s", subscriptionID, azureOutputs.SubscriptionID)
		}
	}

	return nil
}

// Helper methods

func (v *IdentityValidator) getInstallByRunnerGroup(ctx context.Context, runnerGroup *app.RunnerGroup) (*app.Install, error) {
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

func (v *IdentityValidator) getInstallStackWithOutputs(ctx context.Context, installID string) (*app.InstallStack, error) {
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
