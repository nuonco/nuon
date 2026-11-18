package access

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Orgs = authz.DefineResources(permissions.KindOrg, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT id, id AS org_id, ''::text AS app_id, ''::text AS install_id
		FROM orgs WHERE id = ? AND deleted_at = 0`, orgID)
})
