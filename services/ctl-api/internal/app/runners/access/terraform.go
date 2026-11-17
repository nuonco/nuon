package access

import (
	"context"
	"strings"

	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	componentaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TerraformWorkspaces(db *gorm.DB, orgID string) (*authz.Resources, error) {
	owners, err := terraformOwners(db, orgID, "")
	if err != nil {
		return nil, err
	}
	query := db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT w.id, w.org_id, o.app_id, o.install_id FROM terraform_workspaces w
		JOIN (?) o ON o.id = w.owner_id AND o.owner_type = w.owner_type AND o.org_id = w.org_id WHERE w.deleted_at = 0`, owners)
	return authz.NewResources(query, permissions.KindOperation, orgID)
}

func TerraformOwnerTarget(ctx context.Context, db *gorm.DB, orgID, ownerType, ownerID string) (permissions.Target, error) {
	if ownerType == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	owners, err := terraformOwners(db, orgID, ownerType)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := authz.NewResources(owners, permissions.KindOperation, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ownerID)
}

func terraformOwners(db *gorm.DB, orgID, ownerType string) (*gorm.DB, error) {
	components := db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT c.id, c.org_id, i.app_id, i.install_id FROM install_components c
		JOIN (?) i ON i.id = c.install_id AND i.org_id = c.org_id
		JOIN (?) d ON d.id = c.component_id AND d.org_id = c.org_id AND d.app_id = i.app_id WHERE c.deleted_at = 0`, installaccess.Installs.Query(db, orgID), componentaccess.Components.Query(db, orgID))
	sandboxes := db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT s.id, s.org_id, i.app_id, i.install_id FROM install_sandboxes s
		JOIN (?) i ON i.id = s.install_id AND i.org_id = s.org_id WHERE s.deleted_at = 0`, installaccess.Installs.Query(db, orgID))
	queries := []struct {
		kind  string
		query *gorm.DB
	}{
		{"org", orgaccess.Orgs.Query(db, orgID)}, {"orgs", orgaccess.Orgs.Query(db, orgID)},
		{"app", appaccess.Apps.Query(db, orgID)}, {"apps", appaccess.Apps.Query(db, orgID)},
		{"component", componentaccess.Components.Query(db, orgID)}, {"components", componentaccess.Components.Query(db, orgID)},
		{"install", installaccess.Installs.Query(db, orgID)}, {"installs", installaccess.Installs.Query(db, orgID)},
		{"install_components", components}, {"install_sandboxes", sandboxes},
		{"install_sandbox_runs", installaccess.SandboxRuns.Query(db, orgID)},
		{"install_action_workflow_runs", installaccess.ActionRuns.Query(db, orgID)},
		{"install_deploys", installaccess.Deploys.Query(db, orgID)},
		{"install_workflow_steps", installaccess.WorkflowSteps.Query(db, orgID)},
	}
	var parts []string
	var args []any
	for _, owner := range queries {
		if ownerType != "" && ownerType != owner.kind {
			continue
		}
		parts = append(parts, "SELECT ?::text AS owner_type, o.id, o.org_id, o.app_id, o.install_id FROM (?) o")
		args = append(args, owner.kind, owner.query)
	}
	if len(parts) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return db.Session(&gorm.Session{NewDB: true}).Raw(strings.Join(parts, " UNION ALL "), args...), nil
}
