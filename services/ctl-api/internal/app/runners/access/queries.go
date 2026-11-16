package access

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	componentaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var ownerQueries = []struct {
	ownerType string
	query     func(*gorm.DB, string) *gorm.DB
	jobLogs   bool
}{
	{"orgs", orgaccess.Orgs, false},
	{"apps", appaccess.Apps, false},
	{"app_branches", appaccess.Branches, false},
	{"app_branch_configs", appaccess.BranchConfigs, false},
	{"app_branch_runs", appaccess.BranchRuns, false},
	{"app_sandbox_builds", appaccess.SandboxBuilds, true},
	{"installs", installaccess.Installs, false},
	{"install_deploys", installaccess.Deploys, true},
	{"install_action_workflow_runs", installaccess.ActionRuns, true},
	{"install_sandbox_runs", installaccess.SandboxRuns, true},
	{"install_workflows", installaccess.Workflows, false},
	{"install_workflow_steps", installaccess.WorkflowSteps, false},
	{"component_builds", componentaccess.Builds, true},
	{"runners", runners, false},
	{"runner_operations", operations, false},
}

func Jobs(db *gorm.DB, orgID string) (*authz.Resources, error) {
	return jobResources(db, orgID, "")
}

func LogStreams(db *gorm.DB, orgID string) (*authz.Resources, error) {
	return logResources(db, orgID, "")
}

func JobTarget(ctx context.Context, db *gorm.DB, job app.RunnerJob) (permissions.Target, error) {
	if job.DeletedAt != 0 || job.OwnerType == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	resources, err := jobResources(db, job.OrgID, job.OwnerType)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, job.ID)
}

func LogStreamTarget(ctx context.Context, db *gorm.DB, stream app.LogStream) (permissions.Target, error) {
	if stream.DeletedAt != 0 || stream.OwnerType == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	resources, err := logResources(db, stream.OrgID, stream.OwnerType)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, stream.ID)
}

func jobResources(db *gorm.DB, orgID, ownerType string) (*authz.Resources, error) {
	owners, err := owners(db, orgID, ownerType)
	if err != nil {
		return nil, err
	}
	targets := db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT j.id, j.org_id, o.app_id, o.install_id
		FROM runner_jobs j JOIN (?) o ON o.id = j.owner_id AND o.owner_type = j.owner_type AND o.org_id = j.org_id
		WHERE j.deleted_at = 0`, owners)
	return authz.NewResources(targets, permissions.KindRunnerJob, orgID)
}

func logResources(db *gorm.DB, orgID, ownerType string) (*authz.Resources, error) {
	owners, err := owners(db, orgID, ownerType)
	if err != nil {
		return nil, err
	}
	targets := db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT s.id, s.org_id, o.app_id, o.install_id,
		(SELECT COALESCE(jsonb_agg(j.id ORDER BY j.id), '[]'::jsonb) FROM runner_jobs j
		 WHERE j.log_stream_id = s.id AND j.org_id = s.org_id AND j.deleted_at = 0
		 AND j.owner_type = s.owner_type AND j.owner_id = s.owner_id AND o.job_logs) AS related_job_ids
		FROM log_streams s JOIN (?) o ON o.id = s.owner_id AND o.owner_type = s.owner_type AND o.org_id = s.org_id
		WHERE s.deleted_at = 0`, owners)
	return authz.NewResources(targets, permissions.KindLogStream, orgID)
}

func owners(db *gorm.DB, orgID, ownerType string) (*gorm.DB, error) {
	parts := make([]string, 0, len(ownerQueries))
	args := make([]any, 0, len(ownerQueries)*3)
	for _, owner := range ownerQueries {
		if ownerType != "" && ownerType != owner.ownerType {
			continue
		}
		parts = append(parts, "SELECT ?::text AS owner_type, ?::boolean AS job_logs, o.id, o.org_id, o.app_id, o.install_id FROM (?) o")
		args = append(args, owner.ownerType, owner.jobLogs, owner.query(db, orgID))
	}
	if len(parts) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return db.Session(&gorm.Session{NewDB: true}).Raw(strings.Join(parts, " UNION ALL "), args...), nil
}

func runners(db *gorm.DB, orgID string) *gorm.DB {
	db = db.Session(&gorm.Session{NewDB: true})
	owners := db.Raw(`SELECT 'installs'::text AS owner_type, i.id, i.org_id, i.app_id, i.install_id FROM (?) i
		UNION ALL SELECT 'orgs', o.id, o.org_id, o.app_id, o.install_id FROM (?) o`, installaccess.Installs(db, orgID), orgaccess.Orgs(db, orgID))
	return db.Raw(`SELECT r.id, r.org_id, o.app_id, o.install_id FROM runners r
		JOIN runner_groups g ON g.id = r.runner_group_id AND g.org_id = r.org_id AND g.deleted_at = 0
		JOIN (?) o ON o.id = g.owner_id AND o.owner_type = g.owner_type AND o.org_id = g.org_id
		WHERE r.deleted_at = 0`, owners)
}

func operations(db *gorm.DB, orgID string) *gorm.DB {
	return db.Session(&gorm.Session{NewDB: true}).Raw(`SELECT o.id, o.org_id, r.app_id, r.install_id FROM runner_operations o
		JOIN (?) r ON r.id = o.runner_id AND r.org_id = o.org_id WHERE o.deleted_at = 0`, runners(db, orgID))
}
