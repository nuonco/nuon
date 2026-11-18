package access

import (
	"gorm.io/gorm"

	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Notebooks = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT n.id, n.org_id, i.app_id, i.install_id FROM notebooks n
		JOIN (?) i ON i.id = n.install_id AND i.org_id = n.org_id WHERE n.deleted_at = 0`, installaccess.Installs.Query(db, orgID))
})

var Cells = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, n.app_id, n.install_id FROM notebook_cells c
		JOIN (?) n ON n.id = c.notebook_id AND n.org_id = c.org_id WHERE c.deleted_at = 0`, Notebooks.Query(db, orgID))
})

var Runs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, n.app_id, n.install_id FROM notebook_cell_runs r
		JOIN (?) n ON n.id = r.notebook_id AND n.org_id = r.org_id AND n.install_id = r.install_id
		WHERE r.deleted_at = 0`, Notebooks.Query(db, orgID))
})
