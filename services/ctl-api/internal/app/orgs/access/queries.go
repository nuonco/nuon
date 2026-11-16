package access

import "gorm.io/gorm"

func Orgs(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT id, id AS org_id, ''::text AS app_id, ''::text AS install_id
		FROM orgs WHERE id = ? AND deleted_at = 0`, orgID)
}
