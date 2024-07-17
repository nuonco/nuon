package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/plugin/soft_delete"
)

func (m *Migrations) migration061UniqueOrg(ctx context.Context) error {
	// delete org installer that are not the latest update
	var installers []*app.Installer
	res := m.db.WithContext(ctx).Order("created_at desc").Find(&installers)
	if res.Error != nil {
		return res.Error
	}

	installerOrgs := make(map[string]struct{},0)
	for _, installer := range installers {
		_, ok := installerOrgs[installer.OrgID]
		if !ok {
			fmt.Println("visit org id", installer.OrgID)
			installerOrgs[installer.OrgID] = struct{}{}
			continue
		}
		if ok {
			fmt.Println("delete installer", installer.ID)
			res = m.db.WithContext(ctx).Delete(installer)
			if res.Error != nil {
				return res.Error
			}
		}
	}

	// some installers were batch deleted, so we need make their deleted_at unique
	res = m.db.Unscoped().WithContext(ctx).Order("created_at desc").Find(&installers)
	if res.Error != nil {
		return res.Error
	}

	duplicateDeletedAt := make(map[string]uint,0)
	for _, installer := range installers {
		if installer.DeletedAt == 0 {
			continue
		}

		key := fmt.Sprintf(installer.OrgID, installer.DeletedAt)
		fmt.Println("key", key)
		_, ok := duplicateDeletedAt[key]
		if ok {
			duplicateDeletedAt[key] += 1
			fmt.Println("stagger installer deleted_at", installer.ID)
			// stagger the deleted_at to make it unique
			installer.DeletedAt = soft_delete.DeletedAt(uint(installer.DeletedAt) + duplicateDeletedAt[key])
			res = m.db.WithContext(ctx).Save(installer)
			if res.Error != nil {
				return res.Error
			}
			continue
		} else {
			duplicateDeletedAt[key] = 1
		}
	}

	// apply unique constraint on org_id and deleted_at
	addUniqueSQL := `ALTER TABLE installers ADD CONSTRAINT idx_installers_org_id UNIQUE (org_id, deleted_at);`
	if res := m.db.WithContext(ctx).Exec(addUniqueSQL); res.Error != nil {
		return res.Error
	}

	return nil
}
