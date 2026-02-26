package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration094GCPIAMPermissions(ctx context.Context, db *gorm.DB) error {
	alterRoles := `
		ALTER TABLE app_aws_iam_role_configs
			ADD COLUMN IF NOT EXISTS cloud_platform varchar DEFAULT 'aws';
	`
	if res := db.WithContext(ctx).Exec(alterRoles); res.Error != nil {
		return res.Error
	}

	alterPolicies := `
		ALTER TABLE app_aws_iam_policy_configs
			ADD COLUMN IF NOT EXISTS gcp_permissions jsonb DEFAULT '[]';
	`
	if res := db.WithContext(ctx).Exec(alterPolicies); res.Error != nil {
		return res.Error
	}

	return nil
}
