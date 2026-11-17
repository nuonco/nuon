package access

import (
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func Resources(db *gorm.DB, orgID string) (*authz.Resources, error) {
	return Installs.In(db, orgID)
}

var Installs = authz.DefineResources(permissions.KindInstall, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT i.id, i.org_id, i.app_id, i.id AS install_id FROM installs i
		JOIN (?) a ON a.id = i.app_id AND a.org_id = i.org_id WHERE i.deleted_at = 0`, appaccess.Apps.Query(db, orgID))
})

var Deploys = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT d.id, d.org_id, i.app_id, i.install_id FROM install_deploys d
		JOIN install_components c ON c.id = d.install_component_id AND c.org_id = d.org_id AND c.deleted_at = 0
		JOIN (?) i ON i.id = c.install_id AND i.org_id = c.org_id WHERE d.deleted_at = 0`, Installs.Query(db, orgID))
})

var ActionRuns = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, i.app_id, i.install_id FROM install_action_workflow_runs r
		JOIN (?) i ON i.id = r.install_id AND i.org_id = r.org_id WHERE r.deleted_at = 0`, Installs.Query(db, orgID))
})

var SandboxRuns = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT r.id, r.org_id, i.app_id, i.install_id FROM install_sandbox_runs r
		JOIN (?) i ON i.id = r.install_id AND i.org_id = r.org_id WHERE r.deleted_at = 0`, Installs.Query(db, orgID))
})

var Workflows = authz.DefineResources(permissions.KindWorkflow, func(db *gorm.DB, orgID string) *gorm.DB {
	db = db.Session(&gorm.Session{NewDB: true})
	owners := db.Raw(`SELECT 'installs'::text AS owner_type, i.id, i.org_id, i.app_id, i.install_id FROM (?) i
		UNION ALL SELECT 'orgs', o.id, o.org_id, o.app_id, o.install_id FROM (?) o
		UNION ALL SELECT 'apps', a.id, a.org_id, a.app_id, a.install_id FROM (?) a
		UNION ALL SELECT 'app_branches', b.id, b.org_id, b.app_id, b.install_id FROM (?) b
		UNION ALL SELECT 'app_branch_configs', c.id, c.org_id, c.app_id, c.install_id FROM (?) c`,
		Installs.Query(db, orgID), orgaccess.Orgs.Query(db, orgID), appaccess.Apps.Query(db, orgID),
		appaccess.Branches.Query(db, orgID), appaccess.BranchConfigs.Query(db, orgID))
	return db.Raw(`SELECT w.id, w.org_id, o.app_id, o.install_id FROM install_workflows w
		JOIN (?) o ON o.id = w.owner_id AND o.owner_type = w.owner_type AND o.org_id = w.org_id
		WHERE w.deleted_at = 0`, owners)
})

var WorkflowSteps = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT s.id, s.org_id, w.app_id, w.install_id FROM install_workflow_steps s
		JOIN (?) w ON w.id = s.install_workflow_id AND w.org_id = s.org_id WHERE s.deleted_at = 0`, Workflows.Query(db, orgID))
})
