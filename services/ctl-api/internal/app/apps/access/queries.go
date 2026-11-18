package access

import (
	"gorm.io/gorm"

	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Apps = authz.DefineResources(permissions.KindApp, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT a.id, a.org_id, a.id AS app_id, ''::text AS install_id
		FROM apps a JOIN (?) o ON o.id = a.org_id WHERE a.deleted_at = 0`, orgaccess.Orgs.Query(db, orgID))
})

var Branches = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT b.id, b.org_id, a.app_id, a.install_id FROM app_branches b
		JOIN (?) a ON a.id = b.app_id AND a.org_id = b.org_id WHERE b.deleted_at = 0`, Apps.Query(db, orgID))
})

var BranchConfigs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, b.app_id, b.install_id FROM app_branch_configs c
		JOIN (?) b ON b.id = c.app_branch_id AND b.org_id = c.org_id WHERE c.deleted_at = 0`, Branches.Query(db, orgID))
})

var BranchRuns = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, b.app_id, b.install_id FROM app_branch_runs r
		JOIN (?) b ON b.id = r.app_branch_id AND b.org_id = r.org_id WHERE r.deleted_at = 0`, Branches.Query(db, orgID))
})

var InstallGroupRuns = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT g.id, g.org_id, r.app_id, r.install_id FROM install_group_runs g
		JOIN (?) r ON r.id = g.app_branch_run_id AND r.org_id = g.org_id WHERE g.deleted_at = 0`, BranchRuns.Query(db, orgID))
})

var SandboxBuilds = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT b.id, b.org_id, a.app_id, a.install_id FROM app_sandbox_builds b
		JOIN (?) a ON a.id = b.app_id AND a.org_id = b.org_id WHERE b.deleted_at = 0`, Apps.Query(db, orgID))
})
