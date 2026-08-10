package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

const ancestryBackfillBatchSize = 5000

// ancestryBackfill fills the denormalized install_id/app_id ancestry columns
// for one (table, owner_type) pair. installID/appID are SQL expressions valid
// under the joins (or the literal NULL for tiers that don't apply).
type ancestryBackfill struct {
	table     string
	ownerType string
	joins     string
	installID string
	appID     string
}

const (
	joinInstallOwner     = `JOIN installs i ON i.id = t.owner_id`
	joinAppOwner         = `JOIN apps a ON a.id = t.owner_id`
	joinAppBranchOwner   = `JOIN app_branches ab ON ab.id = t.owner_id`
	joinComponentOwner   = `JOIN components c ON c.id = t.owner_id`
	joinDeployOwner      = `JOIN install_deploys d ON d.id = t.owner_id JOIN install_components ic ON ic.id = d.install_component_id JOIN installs i ON i.id = ic.install_id`
	joinSandboxRunOwner  = `JOIN install_sandbox_runs sr ON sr.id = t.owner_id JOIN installs i ON i.id = sr.install_id`
	joinActionRunOwner   = `JOIN install_action_workflow_runs ar ON ar.id = t.owner_id JOIN installs i ON i.id = ar.install_id`
	joinWorkflowStep     = `JOIN install_workflow_steps s ON s.id = t.owner_id JOIN install_workflows w ON w.id = s.install_workflow_id`
	joinComponentBuild   = `JOIN component_builds cb ON cb.id = t.owner_id JOIN component_config_connections ccc ON ccc.id = cb.component_config_connection_id JOIN components c ON c.id = ccc.component_id`
	joinAppSandboxBuild  = `JOIN app_sandbox_builds asb ON asb.id = t.owner_id`
	joinAppBranchRun     = `JOIN app_branch_runs br ON br.id = t.owner_id JOIN app_branches ab ON ab.id = br.app_branch_id`
	joinInstallRunner    = `JOIN runners r ON r.id = t.owner_id JOIN runner_groups g ON g.id = r.runner_group_id AND g.owner_type = 'installs' JOIN installs i ON i.id = g.owner_id`
	joinInstallComponent = `JOIN install_components ic ON ic.id = t.owner_id JOIN installs i ON i.id = ic.install_id`
	joinInstallSandbox   = `JOIN install_sandboxes isb ON isb.id = t.owner_id JOIN installs i ON i.id = isb.install_id`
)

// ancestryBackfills covers every owner_type observed at the tables' creation
// sites. Owner types with no install/app tier (orgs, vcs_connections,
// onboardings, empty/admin-created, runner-supplied unknowns, org-owned
// runner groups) are deliberately absent: their rows keep NULL columns,
// which authorization reads as the org tier.
//
// install_workflows fills first — runner_jobs and log_streams owned by
// install_workflow_steps read the workflow's fresh columns.
var ancestryBackfills = []ancestryBackfill{
	{"install_workflows", "installs", joinInstallOwner, "i.id", "i.app_id"},
	{"install_workflows", "apps", joinAppOwner, "NULL", "a.id"},
	{"install_workflows", "app_branches", joinAppBranchOwner, "NULL", "ab.app_id"},

	{"runner_jobs", "install_deploys", joinDeployOwner, "i.id", "i.app_id"},
	{"runner_jobs", "install_sandbox_runs", joinSandboxRunOwner, "i.id", "i.app_id"},
	{"runner_jobs", "install_action_workflow_runs", joinActionRunOwner, "i.id", "i.app_id"},
	{"runner_jobs", "install_workflow_steps", joinWorkflowStep, "w.install_id", "w.app_id"},
	{"runner_jobs", "component_builds", joinComponentBuild, "NULL", "c.app_id"},
	{"runner_jobs", "app_sandbox_builds", joinAppSandboxBuild, "NULL", "asb.app_id"},
	{"runner_jobs", "runners", joinInstallRunner, "i.id", "i.app_id"},

	{"log_streams", "install_deploys", joinDeployOwner, "i.id", "i.app_id"},
	{"log_streams", "install_sandbox_runs", joinSandboxRunOwner, "i.id", "i.app_id"},
	{"log_streams", "install_action_workflow_runs", joinActionRunOwner, "i.id", "i.app_id"},
	{"log_streams", "install_workflow_steps", joinWorkflowStep, "w.install_id", "w.app_id"},
	{"log_streams", "component_builds", joinComponentBuild, "NULL", "c.app_id"},
	{"log_streams", "app_sandbox_builds", joinAppSandboxBuild, "NULL", "asb.app_id"},
	{"log_streams", "app_branch_runs", joinAppBranchRun, "NULL", "ab.app_id"},
	{"log_streams", "runners", joinInstallRunner, "i.id", "i.app_id"},
	{"log_streams", "runner_operations", joinInstallRunner, "i.id", "i.app_id"},

	{"queues", "installs", joinInstallOwner, "i.id", "i.app_id"},
	{"queues", "apps", joinAppOwner, "NULL", "a.id"},
	{"queues", "app_branches", joinAppBranchOwner, "NULL", "ab.app_id"},
	{"queues", "components", joinComponentOwner, "NULL", "c.app_id"},
	{"queues", "notebooks", `JOIN notebooks n ON n.id = t.owner_id JOIN installs i ON i.id = n.install_id`, "i.id", "i.app_id"},
	{"queues", "runners", joinInstallRunner, "i.id", "i.app_id"},

	{"terraform_workspaces", "install_components", joinInstallComponent, "i.id", "i.app_id"},
	{"terraform_workspaces", "install_sandboxes", joinInstallSandbox, "i.id", "i.app_id"},
}

// Migration126DenormalizeOwnerAncestry backfills the install_id/app_id
// ancestry columns added to the five polymorphic-owner tables. New rows are
// populated by the models' BeforeCreate hooks (ResolveOwnerAncestry); this
// fills everything created before the hooks shipped. Rows written by
// not-yet-updated pods during the deploy rollout stay NULL (org tier) — an
// accepted, fail-closed window.
//
// Each pair drains in batches keyed on still-NULL columns, so the migration
// is idempotent and resumable. The COALESCE guard keeps rows that legitimately
// resolve to (NULL, NULL) out of the batch, guaranteeing termination.
func (m *Migrations) Migration126DenormalizeOwnerAncestry(ctx context.Context, db *gorm.DB) error {
	for _, b := range ancestryBackfills {
		stmt := fmt.Sprintf(`WITH batch AS (
	SELECT t.id AS id, %s AS install_id, %s AS app_id
	FROM %s t
	%s
	WHERE t.owner_type = '%s'
	  AND t.install_id IS NULL AND t.app_id IS NULL
	  AND COALESCE(%s, %s) IS NOT NULL
	LIMIT %d
)
UPDATE %s u
SET install_id = b.install_id, app_id = b.app_id
FROM batch b
WHERE u.id = b.id;`,
			b.installID, b.appID, b.table, b.joins, b.ownerType,
			b.installID, b.appID, ancestryBackfillBatchSize, b.table)

		for {
			res := db.WithContext(ctx).Exec(stmt)
			if res.Error != nil {
				return fmt.Errorf("backfill %s/%s: %w", b.table, b.ownerType, res.Error)
			}
			if res.RowsAffected == 0 {
				break
			}
		}
	}

	return nil
}
