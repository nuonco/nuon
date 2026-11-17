package access

import (
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	componentaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	notebookaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/access"
	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	runneraccess "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var Queues = authz.DefineResources(permissions.KindOperation, func(db *gorm.DB, orgID string) *gorm.DB {
	db = db.Session(&gorm.Session{NewDB: true})
	owners := db.Raw(`SELECT 'orgs'::text AS owner_type, o.* FROM (?) o
		UNION ALL SELECT 'apps', a.* FROM (?) a
		UNION ALL SELECT 'installs', i.* FROM (?) i
		UNION ALL SELECT 'app_branches', b.* FROM (?) b
		UNION ALL SELECT 'components', c.* FROM (?) c
		UNION ALL SELECT 'notebooks', n.* FROM (?) n
		UNION ALL SELECT 'runners', r.* FROM (?) r
		UNION ALL SELECT 'runner_processes', p.id, p.org_id, r.app_id, r.install_id
			FROM runner_processes p JOIN (?) r ON r.id = p.runner_id AND r.org_id = p.org_id WHERE p.deleted_at = 0
		UNION ALL SELECT 'cloud_connections', c.id, c.org_id, o.app_id, o.install_id
			FROM cloud_connections c JOIN (?) o ON o.id = c.org_id WHERE c.deleted_at = 0
		UNION ALL SELECT 'vcs_connections', v.id, v.org_id, o.app_id, o.install_id
			FROM vcs_connections v JOIN (?) o ON o.id = v.org_id WHERE v.deleted_at = 0
		UNION ALL SELECT 'vcs_webhook_subscriptions', s.id, s.org_id, o.app_id, o.install_id
			FROM vcs_webhook_subscriptions s
			JOIN vcs_connections v ON v.id = s.vcs_connection_id AND v.org_id = s.org_id AND v.deleted_at = 0
			JOIN (?) o ON o.id = s.org_id WHERE s.deleted_at = 0`,
		orgaccess.Orgs.Query(db, orgID), appaccess.Apps.Query(db, orgID), installaccess.Installs.Query(db, orgID),
		appaccess.Branches.Query(db, orgID), componentaccess.Components.Query(db, orgID), notebookaccess.Notebooks.Query(db, orgID),
		runneraccess.Runners.Query(db, orgID), runneraccess.Runners.Query(db, orgID), orgaccess.Orgs.Query(db, orgID),
		orgaccess.Orgs.Query(db, orgID), orgaccess.Orgs.Query(db, orgID))
	return db.Raw(`SELECT q.id, q.org_id, o.app_id, o.install_id FROM queues q
		JOIN (?) o ON o.id = q.owner_id AND o.owner_type = q.owner_type AND o.org_id = q.org_id
		WHERE q.deleted_at = 0`, owners)
})
