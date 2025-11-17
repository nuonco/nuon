package migrations

import (
	"context"

	"gorm.io/gorm"
)

func (m *Migrations) Migration092BackfillOrgID(ctx context.Context, db *gorm.DB) error {
	updateTerraformStateJson := `UPDATE terraform_workspace_state_jsons
SET org_id = terraform_workspaces.org_id
FROM terraform_workspaces
WHERE terraform_workspace_state_jsons.workspace_id = terraform_workspaces.id;
`

	updateTerraformStateLock := `UPDATE terraform_workspace_locks
SET org_id = terraform_workspaces.org_id
FROM terraform_workspaces
WHERE terraform_workspace_locks.workspace_id = terraform_workspaces.id;
`

	if res := db.WithContext(ctx).
		Exec(updateTerraformStateJson); res.Error != nil {
		return res.Error
	}

	if res := db.WithContext(ctx).
		Exec(updateTerraformStateLock); res.Error != nil {
		return res.Error
	}

	return nil
}
