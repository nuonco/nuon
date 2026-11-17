package access

import (
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Actions = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT w.id, w.org_id, a.app_id, a.install_id FROM action_workflows w
		JOIN (?) a ON a.id = w.app_id AND a.org_id = w.org_id WHERE w.deleted_at = 0`, appaccess.Apps.Query(db, orgID))
})

var Configs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, w.app_id, w.install_id FROM action_workflow_configs c
		JOIN (?) w ON w.id = c.action_workflow_id AND w.org_id = c.org_id AND w.app_id = c.app_id
		WHERE c.deleted_at = 0`, Actions.Query(db, orgID))
})

var InstallActions = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT w.id, w.org_id, i.app_id, i.install_id FROM install_action_workflows w
		JOIN (?) i ON i.id = w.install_id AND i.org_id = w.org_id
		JOIN (?) a ON a.id = w.action_workflow_id AND a.org_id = w.org_id AND a.app_id = i.app_id
		WHERE w.deleted_at = 0`, installaccess.Installs.Query(db, orgID), Actions.Query(db, orgID))
})

var Runs = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return installaccess.ActionRuns.Query(db, orgID)
})

var Steps = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT s.id, s.org_id, r.app_id, r.install_id FROM install_action_workflow_run_steps s
		JOIN (?) r ON r.id = s.install_action_workflow_run_id AND r.org_id = s.org_id WHERE s.deleted_at = 0`, Runs.Query(db, orgID))
})
