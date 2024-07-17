package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/plugin/soft_delete"
)

func (m *Migrations) migration062InstallerUniqueOrg(ctx context.Context) error {
	// delete org installer that are not the latest update
	var i1 []*app.Installer
	res := m.db.WithContext(ctx).Order("updated_at DESC").Find(&i1)
	if res.Error != nil {
		return res.Error
	}

	installerOrgs := make(map[string]struct{}, 0)
	for _, installer := range i1 {
		_, ok := installerOrgs[installer.OrgID]
		if !ok {
			fmt.Println("visit org ID=", installer.OrgID)
			installerOrgs[installer.OrgID] = struct{}{}
			continue
		}
		if ok {
			fmt.Println("delete installer ID=", installer.ID)
			res = m.db.WithContext(ctx).Delete(installer)
			if res.Error != nil {
				return res.Error
			}
		}
	}

	var i2 []*app.Installer
	// some installers were batch deleted, so we need make their deleted_at unique
	res = m.db.Unscoped().WithContext(ctx).Where("deleted_at != ?", 0).Order("updated_at DESC").Find(&i2)
	if res.Error != nil {
		return res.Error
	}

	duplicateDeletedAt := make(map[string]uint, 0)
	for _, installer := range i2 {
		i := installer
		if uint(i.DeletedAt) == 0 {
			fmt.Println("skip installer ID=", i.ID)
			continue
		}

		key := fmt.Sprintf(i.OrgID, installer.DeletedAt)
		fmt.Println("key", key)
		_, ok := duplicateDeletedAt[key]
		if ok {
			duplicateDeletedAt[key] += 1
			fmt.Println("stagger installer deleted_at ID=", i.ID)
			// stagger the deleted_at to make it unique
			i.DeletedAt = soft_delete.DeletedAt(uint(i.DeletedAt) + duplicateDeletedAt[key])
			res = m.db.WithContext(ctx).Save(i)
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
