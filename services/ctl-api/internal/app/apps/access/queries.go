package access

import (
	"gorm.io/gorm"

	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
)

func Apps(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT a.id, a.org_id, a.id AS app_id, ''::text AS install_id
		FROM apps a JOIN (?) o ON o.id = a.org_id WHERE a.deleted_at = 0`, orgaccess.Orgs(db, orgID))
}

func Branches(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT b.id, b.org_id, a.app_id, a.install_id FROM app_branches b
		JOIN (?) a ON a.id = b.app_id AND a.org_id = b.org_id WHERE b.deleted_at = 0`, Apps(db, orgID))
}

func BranchConfigs(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, b.app_id, b.install_id FROM app_branch_configs c
		JOIN (?) b ON b.id = c.app_branch_id AND b.org_id = c.org_id WHERE c.deleted_at = 0`, Branches(db, orgID))
}

func BranchRuns(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, b.app_id, b.install_id FROM app_branch_runs r
		JOIN (?) b ON b.id = r.app_branch_id AND b.org_id = r.org_id WHERE r.deleted_at = 0`, Branches(db, orgID))
}

func SandboxBuilds(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT b.id, b.org_id, a.app_id, a.install_id FROM app_sandbox_builds b
		JOIN (?) a ON a.id = b.app_id AND a.org_id = b.org_id WHERE b.deleted_at = 0`, Apps(db, orgID))
}
