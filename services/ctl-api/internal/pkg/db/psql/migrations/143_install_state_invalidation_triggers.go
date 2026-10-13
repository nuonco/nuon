package migrations

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"gorm.io/gorm"

	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

//go:embed 143_install_state_invalidation_triggers.sql
var installStateInvalidationTriggersTemplate string

type installStateInstallColumn struct {
	Column   string
	Partials []pkgstate.PartialName
}

type installStateTarget struct {
	InstallID string
	Joins     string
	Where     string
	Partials  []pkgstate.PartialName
}

type installStateSource struct {
	Table   string
	Alias   string
	Columns []string
	Targets []installStateTarget
}

type installStateJobOwner struct {
	OwnerType string
	Joins     string
	InstallID string
	Partials  []pkgstate.PartialName
}

var installStateInstallColumns = []installStateInstallColumn{
	{Column: "org_id", Partials: []pkgstate.PartialName{pkgstate.PartialOrg}},
	{Column: "app_id", Partials: []pkgstate.PartialName{pkgstate.PartialApp}},
	{Column: "app_config_id", Partials: []pkgstate.PartialName{pkgstate.PartialInputs}},
	{Column: "name", Partials: []pkgstate.PartialName{pkgstate.PartialCloud}},
	{Column: "deleted_at", Partials: pkgstate.AllPartials},
}

var installStateJobOwners = []installStateJobOwner{
	{
		OwnerType: "install_deploys",
		Joins:     "JOIN install_deploys d ON d.id = j.owner_id JOIN install_components ic ON ic.id = d.install_component_id",
		InstallID: "ic.install_id",
		Partials:  []pkgstate.PartialName{pkgstate.PartialComponents},
	},
	{
		OwnerType: "install_sandbox_runs",
		Joins:     "JOIN install_sandbox_runs s ON s.id = j.owner_id",
		InstallID: "s.install_id",
		Partials:  []pkgstate.PartialName{pkgstate.PartialSandbox, pkgstate.PartialDomain},
	},
	{
		OwnerType: "install_action_workflow_runs",
		Joins:     "JOIN install_action_workflow_runs w ON w.id = j.owner_id",
		InstallID: "w.install_id",
		Partials:  []pkgstate.PartialName{pkgstate.PartialActions},
	},
	{
		OwnerType: "install_action_workflow_run_steps",
		Joins:     "JOIN install_action_workflow_run_steps st ON st.id = j.owner_id JOIN install_action_workflow_runs w ON w.id = st.install_action_workflow_run_id",
		InstallID: "w.install_id",
		Partials:  []pkgstate.PartialName{pkgstate.PartialActions},
	},
}

func installStateJobTargets(jobJoin string) []installStateTarget {
	targets := make([]installStateTarget, 0, len(installStateJobOwners))
	for _, owner := range installStateJobOwners {
		targets = append(targets, installStateTarget{
			InstallID: owner.InstallID,
			Joins:     strings.TrimSpace(jobJoin + " " + owner.Joins),
			Where:     fmt.Sprintf("j.owner_type = '%s'", owner.OwnerType),
			Partials:  owner.Partials,
		})
	}
	return append(targets, installStateTarget{
		InstallID: "g.owner_id",
		Joins:     strings.TrimSpace(jobJoin + " JOIN runners rn ON rn.id = j.runner_id JOIN runner_groups g ON g.id = rn.runner_group_id"),
		Where:     "g.owner_type = 'installs'",
		Partials:  []pkgstate.PartialName{pkgstate.PartialSecrets},
	})
}

func installStateDirect(partials ...pkgstate.PartialName) []installStateTarget {
	return []installStateTarget{{InstallID: "r.install_id", Partials: partials}}
}

var installStateSources = []installStateSource{
	{
		Table:   "orgs",
		Columns: []string{"name", "status", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "i.id", Joins: "JOIN installs i ON i.org_id = r.id", Partials: []pkgstate.PartialName{pkgstate.PartialOrg}}},
	},
	{
		Table:   "apps",
		Columns: []string{"name", "status", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "i.id", Joins: "JOIN installs i ON i.app_id = r.id", Partials: []pkgstate.PartialName{pkgstate.PartialApp}}},
	},
	{
		Table:   "app_secrets",
		Columns: []string{"app_id", "name", "value", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "i.id", Joins: "JOIN installs i ON i.app_id = r.app_id", Partials: []pkgstate.PartialName{pkgstate.PartialApp}}},
	},
	{
		Table:   "runner_groups",
		Columns: []string{"owner_id", "owner_type", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "r.owner_id", Where: "r.owner_type = 'installs'", Partials: []pkgstate.PartialName{pkgstate.PartialRunner, pkgstate.PartialSecrets}}},
	},
	{
		Table:   "runners",
		Columns: []string{"runner_group_id", "status", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "g.owner_id", Joins: "JOIN runner_groups g ON g.id = r.runner_group_id", Where: "g.owner_type = 'installs'", Partials: []pkgstate.PartialName{pkgstate.PartialRunner, pkgstate.PartialSecrets}}},
	},
	{
		Table:   "aws_accounts",
		Columns: []string{"install_id", "region", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialCloud),
	},
	{
		Table:   "azure_accounts",
		Columns: []string{"install_id", "location", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialCloud),
	},
	{
		Table:   "gcp_accounts",
		Columns: []string{"install_id", "project_id", "region", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialCloud),
	},
	{
		Table:   "app_input_configs",
		Columns: []string{"app_config_id", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "i.id", Joins: "JOIN installs i ON i.app_config_id = r.app_config_id", Partials: []pkgstate.PartialName{pkgstate.PartialInputs}}},
	},
	{
		Table:   "app_inputs",
		Columns: []string{"app_input_config_id", "name", `"default"`, "deleted_at"},
		Targets: []installStateTarget{{InstallID: "i.id", Joins: "JOIN app_input_configs c ON c.id = r.app_input_config_id JOIN installs i ON i.app_config_id = c.app_config_id", Partials: []pkgstate.PartialName{pkgstate.PartialInputs}}},
	},
	{
		Table:   "install_inputs",
		Columns: []string{"install_id", "values", "created_at", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialInputs),
	},
	{
		Table:   "install_sandbox_runs",
		Columns: []string{"install_id", "status", "status_v2->>'status'", "created_at", "app_sandbox_config_id", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "r.install_id", Where: "r.status = 'active'", Partials: []pkgstate.PartialName{pkgstate.PartialSandbox, pkgstate.PartialDomain}}},
	},
	{
		Table:   "app_sandbox_configs",
		Columns: []string{"deleted_at"},
		Targets: []installStateTarget{{InstallID: "s.install_id", Joins: "JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.id", Partials: []pkgstate.PartialName{pkgstate.PartialSandbox}}},
	},
	{
		Table:   "public_git_vcs_configs",
		Columns: []string{"component_config_id", "component_config_type", "directory", "branch", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "s.install_id", Joins: "JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.component_config_id", Where: "r.component_config_type = 'app_sandbox_configs'", Partials: []pkgstate.PartialName{pkgstate.PartialSandbox}}},
	},
	{
		Table:   "connected_github_vcs_configs",
		Columns: []string{"component_config_id", "component_config_type", "directory", "branch", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "s.install_id", Joins: "JOIN install_sandbox_runs s ON s.app_sandbox_config_id = r.component_config_id", Where: "r.component_config_type = 'app_sandbox_configs'", Partials: []pkgstate.PartialName{pkgstate.PartialSandbox}}},
	},
	{
		Table:   "install_stacks",
		Columns: []string{"install_id", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialStack),
	},
	{
		Table:   "install_stack_versions",
		Columns: []string{"install_stack_id", "install_id", "created_at", "quick_link_url", "template_url", "contents", "checksum", "status->>'status'", "deleted_at"},
		Targets: []installStateTarget{
			{InstallID: "r.install_id", Partials: []pkgstate.PartialName{pkgstate.PartialStack}},
			{InstallID: "s.install_id", Joins: "JOIN install_stacks s ON s.id = r.install_stack_id", Partials: []pkgstate.PartialName{pkgstate.PartialStack}},
		},
	},
	{
		Table:   "install_stack_outputs",
		Columns: []string{"install_stack_id", "data", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "s.install_id", Joins: "JOIN install_stacks s ON s.id = r.install_stack_id", Partials: []pkgstate.PartialName{pkgstate.PartialStack}}},
	},
	{
		Table:   "install_components",
		Columns: []string{"install_id", "component_id", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialComponents),
	},
	{
		Table:   "components",
		Columns: []string{"name", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "ic.install_id", Joins: "JOIN install_components ic ON ic.component_id = r.id", Partials: []pkgstate.PartialName{pkgstate.PartialComponents}}},
	},
	{
		Table:   "install_deploys",
		Columns: []string{"install_component_id", "status", "status_v2->>'status'", "component_build_id", "created_at", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "ic.install_id", Joins: "JOIN install_components ic ON ic.id = r.install_component_id", Where: "r.status IN ('active', 'inactive')", Partials: []pkgstate.PartialName{pkgstate.PartialComponents}}},
	},
	{
		Table:   "install_action_workflows",
		Columns: []string{"install_id", "action_workflow_id", "deleted_at"},
		Targets: installStateDirect(pkgstate.PartialActions),
	},
	{
		Table:   "action_workflows",
		Columns: []string{"name", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "w.install_id", Joins: "JOIN install_action_workflows w ON w.action_workflow_id = r.id", Partials: []pkgstate.PartialName{pkgstate.PartialActions}}},
	},
	{
		Table:   "install_action_workflow_runs",
		Columns: []string{"install_id", "install_action_workflow_id", "status", "status_v2->>'status'", "created_at", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "r.install_id", Where: "r.status = 'finished'", Partials: []pkgstate.PartialName{pkgstate.PartialActions}}},
	},
	{
		Table:   "install_action_workflow_run_steps",
		Columns: []string{"install_action_workflow_run_id", "deleted_at"},
		Targets: []installStateTarget{{InstallID: "w.install_id", Joins: "JOIN install_action_workflow_runs w ON w.id = r.install_action_workflow_run_id", Partials: []pkgstate.PartialName{pkgstate.PartialActions}}},
	},
	{
		Table:   "runner_jobs",
		Alias:   "j",
		Columns: []string{"owner_id", "owner_type", "type", "operation", "runner_id", "created_at", "deleted_at"},
		Targets: installStateJobTargets(""),
	},
	{
		Table:   "runner_job_executions",
		Columns: []string{"runner_job_id", "status", "created_at", "deleted_at"},
		Targets: installStateJobTargets("JOIN runner_jobs j ON j.id = r.runner_job_id"),
	},
	{
		Table:   "runner_job_execution_outputs",
		Columns: []string{"runner_job_execution_id", "outputs", "outputs_blob", "deleted_at"},
		Targets: installStateJobTargets("JOIN runner_job_executions e ON e.id = r.runner_job_execution_id JOIN runner_jobs j ON j.id = e.runner_job_id"),
	},
}

func installStateSQLArray(partials []pkgstate.PartialName) string {
	quoted := make([]string, 0, len(partials))
	for _, p := range partials {
		quoted = append(quoted, "'"+string(p)+"'")
	}
	return "ARRAY[" + strings.Join(quoted, ", ") + "]::text[]"
}

func installStateColumns(alias string, columns []string) string {
	prefixed := make([]string, 0, len(columns))
	for _, c := range columns {
		prefixed = append(prefixed, alias+"."+c)
	}
	return strings.Join(prefixed, ", ")
}

func renderInstallStateInvalidationTriggersSQL() (string, error) {
	tmpl, err := template.New("install_state_invalidation").Option("missingkey=error").Funcs(template.FuncMap{
		"array":   installStateSQLArray,
		"columns": installStateColumns,
	}).Parse(installStateInvalidationTriggersTemplate)
	if err != nil {
		return "", err
	}

	sources := make([]installStateSource, 0, len(installStateSources))
	for _, s := range installStateSources {
		if s.Alias == "" {
			s.Alias = "r"
		}
		sources = append(sources, s)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, map[string]any{
		"AllPartials":    pkgstate.AllPartials,
		"InstallColumns": installStateInstallColumns,
		"Sources":        sources,
	}); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (m *Migrations) Migration143InstallStateInvalidationTriggers(ctx context.Context, db *gorm.DB) error {
	sql, err := renderInstallStateInvalidationTriggersSQL()
	if err != nil {
		return err
	}

	if res := db.WithContext(ctx).Exec(sql); res.Error != nil {
		return res.Error
	}

	return nil
}
