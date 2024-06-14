package migrations

import (
	"context"
	"fmt"
)

const (
	migration050SQLTemplate string = `
ALTER TABLE %s DROP CONSTRAINT fk_%s_created_by
`
)

var migration050Tables []string = []string{
	"app_configs",
	"app_input_configs",
	"app_input_groups",
	"app_inputs",
	"app_runner_configs",
	"app_sandbox_configs",
	"app_secrets",
	"apps",
	"aws_accounts",
	"awsecr_image_configs",
	"azure_accounts",
	"component_builds",
	"component_config_connections",
	"component_dependencies",
	"component_release_steps",
	"component_releases",
	"components",
	"connected_github_vcs_configs",
	"docker_build_component_configs",
	"external_image_component_configs",
	"helm_component_configs",
	"install_components",
	"install_deploys",
	"install_events",
	"install_inputs",
	"install_sandbox_runs",
	"installer_metadata",
	"installers",
	"installs",
	"job_component_configs",
	"notifications_configs",
	"org_health_checks",
	"org_invites",
	"orgs",
	"public_git_vcs_configs",
	"sandbox_releases",
	"sandboxes",
	"terraform_module_component_configs",
	"user_orgs",
	"vcs_connection_commits",
	"vcs_connections",
}

func (a *Migrations) migration050UpdateCreatedByConstraints(ctx context.Context) error {
	for _, table := range migration050Tables {
		a.l.Info("migrating " + table)

		sql := fmt.Sprintf(migration050SQLTemplate, table, table)
		if res := a.db.WithContext(ctx).Exec(sql); res.Error != nil {
			return res.Error
		}
	}

	return nil
}
