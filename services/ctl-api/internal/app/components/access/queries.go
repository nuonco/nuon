package access

import (
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
)

func Builds(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT b.id, b.org_id, a.app_id, a.install_id FROM component_builds b
		JOIN component_config_connections cc ON cc.id = b.component_config_connection_id AND cc.org_id = b.org_id AND cc.deleted_at = 0
		JOIN components c ON c.id = cc.component_id AND c.org_id = cc.org_id AND c.deleted_at = 0
		JOIN (?) a ON a.id = c.app_id AND a.org_id = c.org_id WHERE b.deleted_at = 0`, appaccess.Apps(db, orgID))
}
