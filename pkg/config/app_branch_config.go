package config

import (
	"fmt"
	"regexp"

	"github.com/invopop/jsonschema"
)

type AppBranchInstallGroupConfig struct {
	Name  string `mapstructure:"name" toml:"name" jsonschema:"required"`
	Order int    `mapstructure:"order" toml:"order"`

	InstallIDs    []string          `mapstructure:"install_ids,omitempty" toml:"install_ids,omitempty"`
	InstallNames  []string          `mapstructure:"install_names,omitempty" toml:"install_names,omitempty"`
	LabelSelector map[string]string `mapstructure:"label_selector,omitempty" toml:"label_selector,omitempty"`
}

type AppBranchPreviewConfig struct {
	Mode          string            `mapstructure:"mode,omitempty" toml:"mode,omitempty"`
	InstallID     string            `mapstructure:"install_id,omitempty" toml:"install_id,omitempty"`
	InstallName   string            `mapstructure:"install_name,omitempty" toml:"install_name,omitempty"`
	LabelSelector map[string]string `mapstructure:"label_selector,omitempty" toml:"label_selector,omitempty"`
	SetStatuses   *bool             `mapstructure:"set_statuses,omitempty" toml:"set_statuses,omitempty"`
	Comment       *bool             `mapstructure:"comment,omitempty" toml:"comment,omitempty"`
}

func (c AppBranchPreviewConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "mode", "preview run mode: plan-only, apply, or build-only")
	addDescription(schema, "install_id", "default install ID for preview runs")
	addDescription(schema, "install_name", "default install name for preview runs, resolved to an ID at sync time")
	addDescription(schema, "label_selector", "label key-value pairs to select the default preview install")
	addDescription(schema, "set_statuses", "whether to set GitHub commit statuses for preview runs")
	addDescription(schema, "comment", "whether to comment on the pull request with preview results")
}

func (c AppBranchInstallGroupConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install group")
	addDescription(schema, "order", "deployment order (lower runs first)")
	addDescription(schema, "install_ids", "static list of install IDs")
	addDescription(schema, "install_names", "static list of install names, resolved to IDs at sync time")
	addDescription(schema, "label_selector", "label key-value pairs to dynamically match installs")
}

type AppBranchConfig struct {
	Name          string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`

	InstallGroups []AppBranchInstallGroupConfig `mapstructure:"install_groups,omitempty" toml:"install_groups,omitempty"`

	Preview *AppBranchPreviewConfig `mapstructure:"preview,omitempty" toml:"preview,omitempty"`

	PostDeployRunbooks []string `mapstructure:"post_deploy_runbooks,omitempty" toml:"post_deploy_runbooks,omitempty" json:"post_deploy_runbooks,omitempty"`

	IgnoreChangesRegex string `mapstructure:"ignore_changes_regex,omitempty" toml:"ignore_changes_regex,omitempty" json:"ignore_changes_regex,omitempty"`

	SendStatusesOnIgnore bool `mapstructure:"send_statuses_on_ignore,omitempty" toml:"send_statuses_on_ignore,omitempty" json:"send_statuses_on_ignore,omitempty"`
}

func (c AppBranchConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the app branch")
	addDescription(schema, "connected_repo", "connected GitHub repo the branch tracks")
	addDescription(schema, "public_repo", "public git repo the branch tracks")
	addDescription(schema, "install_groups", "ordered deployment groups for this branch")
	addDescription(schema, "preview", "default preview run settings for this branch")
	addDescription(schema, "post_deploy_runbooks", "names of runbooks to run on each install, in order, after its deploy succeeds; resolved to IDs at sync time")
	addDescription(schema, "ignore_changes_regex", "RE2 regex matched against every changed file path; a run whose entire changed file set matches is not attempted")
	addDescription(schema, "send_statuses_on_ignore", "whether to send a successful commit status when a run is ignored by ignore_changes_regex")
}

func (c *AppBranchConfig) Validate() error {
	if c.IgnoreChangesRegex != "" {
		if _, err := regexp.Compile(c.IgnoreChangesRegex); err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: ignore_changes_regex is not a valid regular expression: %v", c.Name, err),
			}
		}
	}

	for _, name := range c.PostDeployRunbooks {
		if name == "" {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: post_deploy_runbooks entries must be non-empty runbook names", c.Name),
			}
		}
	}

	for _, g := range c.InstallGroups {
		hasStatic := len(g.InstallIDs) > 0 || len(g.InstallNames) > 0
		hasLabels := len(g.LabelSelector) > 0
		if hasStatic && hasLabels {
			return ErrConfig{
				Description: fmt.Sprintf("install group %q: label_selector is mutually exclusive with install_ids and install_names", g.Name),
			}
		}
	}
	if c.Preview != nil {
		hasInstallID := c.Preview.InstallID != ""
		hasInstallName := c.Preview.InstallName != ""
		hasLabels := len(c.Preview.LabelSelector) > 0
		if hasInstallID && hasLabels {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: preview.label_selector is mutually exclusive with install_id", c.Name),
			}
		}
		if hasInstallName && hasLabels {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: preview.label_selector is mutually exclusive with install_name", c.Name),
			}
		}
		if hasInstallID && hasInstallName {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: preview.install_id is mutually exclusive with install_name", c.Name),
			}
		}
		mode := c.Preview.Mode
		if mode == "" {
			mode = "plan-only"
		}
		if mode != "build-only" && !hasInstallID && !hasInstallName && !hasLabels {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: preview requires install_id, install_name, or label_selector for mode %q", c.Name, mode),
			}
		}
	}
	return nil
}

func (c *AppBranchConfig) parse() error {
	return nil
}
