package access

import (
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Runbooks = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, a.app_id, a.install_id FROM runbooks r
		JOIN (?) a ON a.id = r.app_id AND a.org_id = r.org_id WHERE r.deleted_at = 0`, appaccess.Apps.Query(db, orgID))
})

var Configs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, r.app_id, r.install_id FROM runbook_configs c
		JOIN (?) r ON r.id = c.runbook_id AND r.org_id = c.org_id AND r.app_id = c.app_id WHERE c.deleted_at = 0`, Runbooks.Query(db, orgID))
})

var InstallRunbooks = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, i.app_id, i.install_id FROM install_runbooks r
		JOIN (?) i ON i.id = r.install_id AND i.org_id = r.org_id
		JOIN (?) b ON b.id = r.runbook_id AND b.org_id = r.org_id AND b.app_id = i.app_id
		WHERE r.deleted_at = 0`, installaccess.Installs.Query(db, orgID), Runbooks.Query(db, orgID))
})

var Runs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, i.app_id, i.install_id FROM install_runbook_runs r
		JOIN (?) i ON i.id = r.install_id AND i.org_id = r.org_id WHERE r.deleted_at = 0`, installaccess.Installs.Query(db, orgID))
})
