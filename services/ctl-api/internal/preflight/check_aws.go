package preflight

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

var awsCheck = Check{
	Name:        "aws",
	Description: "management account role assumption",

	// The management ARNs are validated per-cloud in internal.NewConfig, so on a
	// GCP or Azure control plane they are legitimately empty.
	Skip: func(cfg *internal.Config) (string, bool) {
		if !cfg.IsAWS() {
			return "cloud_provider=" + cfg.CloudProvider, true
		}

		return "", false
	},

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "management_iam_role_arn", Value: cfg.ManagementIAMRoleARN, Required: true},
			{Name: "management_account_id", Value: cfg.ManagementAccountID, Required: true},
			{Name: "management_ecr_registry_id", Value: cfg.ManagementECRRegistryID, Required: true},
			{Name: "management_ecr_registry_arn", Value: cfg.ManagementECRRegistryARN},
			{Name: "blob_storage_region", Value: cfg.BlobStorageRegion},
		}
	},

	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		awsCfg, err := credentials.Fetch(ctx, &credentials.Config{
			Region: cfg.BlobStorageRegion,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:                cfg.ManagementIAMRoleARN,
				SessionName:            "ctl-api-preflight",
				SessionDurationSeconds: 15 * 60,
			},
		})
		if err != nil {
			return "", fmt.Errorf("unable to assume %s: %w", cfg.ManagementIAMRoleARN, err)
		}

		identity, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return "", fmt.Errorf("sts:GetCallerIdentity failed: %w", err)
		}

		// A valid role in the wrong account is the misconfiguration that a bare
		// GetCallerIdentity would report as healthy.
		account := aws.ToString(identity.Account)
		if account != cfg.ManagementAccountID {
			return "", fmt.Errorf("assumed role is in account %s, expected management_account_id=%s",
				account, cfg.ManagementAccountID)
		}

		return fmt.Sprintf("assumed role %s", summary("account", account,
			"arn", aws.ToString(identity.Arn))), nil
	},
}
