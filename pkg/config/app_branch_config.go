package config

import (
	"fmt"
	"regexp"

	"github.com/invopop/jsonschema"
)

type AppBranchInstallGroupConfig struct {
	Name  string `mapstructure:"name" toml:"name" jsonschema:"required"`
	Order int    `mapstructure:"order" toml:"order"`

	LabelSelector map[string]string `mapstructure:"label_selector,omitempty" toml:"label_selector,omitempty"`
	Default       bool              `mapstructure:"default,omitempty" toml:"default,omitempty"`

	AutoApproveOnPoliciesPassing *bool `mapstructure:"auto_approve_on_policies_passing,omitempty" toml:"auto_approve_on_policies_passing,omitempty"`
}

type AppBranchPreviewConfig struct {
	Mode          string            `mapstructure:"mode,omitempty" toml:"mode,omitempty" jsonschema:"enum=none,enum=plan-only,enum=apply,enum=build-only"`
	InstallID     string            `mapstructure:"install_id,omitempty" toml:"install_id,omitempty"`
	InstallName   string            `mapstructure:"install_name,omitempty" toml:"install_name,omitempty"`
	LabelSelector map[string]string `mapstructure:"label_selector,omitempty" toml:"label_selector,omitempty"`
	SetStatuses   *bool             `mapstructure:"set_statuses,omitempty" toml:"set_statuses,omitempty"`
	Comment       *bool             `mapstructure:"comment,omitempty" toml:"comment,omitempty"`
	IgnoreDrafts  *bool             `mapstructure:"ignore_drafts,omitempty" toml:"ignore_drafts,omitempty"`
	React         *bool             `mapstructure:"react,omitempty" toml:"react,omitempty"`
}

type AppBranchRunConfig struct {
	Mode        string `mapstructure:"mode,omitempty" toml:"mode,omitempty" jsonschema:"enum=push,enum=on_tag,enum=on_github_label,enum=manual_only"`
	TagPrefix   string `mapstructure:"tag_prefix,omitempty" toml:"tag_prefix,omitempty"`
	GithubLabel string `mapstructure:"github_label,omitempty" toml:"github_label,omitempty"`
}

func (c AppBranchRunConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "mode", "automatic run mode: push, on_tag, on_github_label, or manual_only")
	addDescription(schema, "tag_prefix", "case-sensitive git tag prefix required by on_tag")
	addDescription(schema, "github_label", "exact pull request label required by on_github_label")
}

func (c AppBranchPreviewConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "mode", "preview run mode: none, plan-only, apply, or build-only")
	addDescription(schema, "install_id", "default install ID for preview runs")
	addDescription(schema, "install_name", "default install name for preview runs, resolved to an ID at sync time")
	addDescription(schema, "label_selector", "label key-value pairs to select the default preview install")
	addDescription(schema, "set_statuses", "whether to set GitHub commit statuses for preview runs")
	addDescription(schema, "comment", "whether to comment on the pull request with preview results")
	addDescription(schema, "ignore_drafts", "skip preview runs for draft pull requests until they are marked ready for review. Defaults to true")
	addDescription(schema, "react", "whether to add a GitHub eyes reaction when a preview run starts. Defaults to true")
}

func (c AppBranchInstallGroupConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install group")
	addDescription(schema, "order", "deployment order (lower runs first)")
	addDescription(schema, "label_selector", "label key-value pairs to dynamically match installs")
	addDescription(schema, "default", "whether unmatched installs owned by this branch deploy in this group")
	addDescription(schema, "auto_approve_on_policies_passing", "Auto-approve this group's plan when all policy checks pass. Defaults to false")
}

type AppBranchConfig struct {
	Name          string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`

	InstallGroups []AppBranchInstallGroupConfig `mapstructure:"install_groups,omitempty" toml:"install_groups,omitempty"`

	Preview *AppBranchPreviewConfig `mapstructure:"preview,omitempty" toml:"preview,omitempty"`
	Run     *AppBranchRunConfig     `mapstructure:"run,omitempty" toml:"run,omitempty"`

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
	addDescription(schema, "run", "controls which VCS events automatically run this branch")
	addDescription(schema, "post_deploy_runbooks", "names of runbooks to run on each install, in order, after its deploy succeeds; resolved to IDs at sync time")
	addDescription(schema, "ignore_changes_regex", "RE2 regex matched against every changed file path; a run whose entire changed file set matches is not attempted")
	addDescription(schema, "send_statuses_on_ignore", "whether to send a successful commit status when a run is ignored by ignore_changes_regex")
}

func (c *AppBranchConfig) Validate() error {
	if c.Run != nil {
		mode := c.Run.Mode
		if mode == "" || mode == "all" {
			mode = "push"
		} else if mode == "on_tag_prefix" {
			mode = "on_tag"
		}
		switch mode {
		case "push", "manual_only":
			if c.Run.TagPrefix != "" || c.Run.GithubLabel != "" {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode %q cannot set tag_prefix or github_label", c.Name, mode)}
			}
		case "on_tag":
			if c.Run.TagPrefix == "" {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode on_tag requires tag_prefix", c.Name)}
			}
			if c.Run.GithubLabel != "" {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode on_tag cannot set github_label", c.Name)}
			}
		case "on_github_label":
			if c.Run.GithubLabel == "" {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode on_github_label requires github_label", c.Name)}
			}
			if c.Run.TagPrefix != "" {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode on_github_label cannot set tag_prefix", c.Name)}
			}
			if c.ConnectedRepo == nil {
				return ErrConfig{Description: fmt.Sprintf("branch %q: run mode on_github_label requires connected_repo", c.Name)}
			}
		default:
			return ErrConfig{Description: fmt.Sprintf("branch %q: unknown run mode %q (valid modes: push, on_tag, on_github_label, manual_only)", c.Name, mode)}
		}
	}

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

	defaultGroups := 0
	groupNames := make(map[string]struct{}, len(c.InstallGroups))
	for _, g := range c.InstallGroups {
		if _, ok := groupNames[g.Name]; ok {
			return ErrConfig{
				Description: fmt.Sprintf("branch %q: install group names must be unique; %q is duplicated", c.Name, g.Name),
			}
		}
		groupNames[g.Name] = struct{}{}
		if g.Default {
			defaultGroups++
		}
		if g.Default && len(g.LabelSelector) > 0 {
			return ErrConfig{
				Description: fmt.Sprintf("install group %q: default is mutually exclusive with label_selector", g.Name),
			}
		}
		if !g.Default && len(g.LabelSelector) == 0 {
			return ErrConfig{
				Description: fmt.Sprintf("install group %q: either default or label_selector is required", g.Name),
			}
		}
	}
	if defaultGroups > 1 {
		return ErrConfig{
			Description: fmt.Sprintf("branch %q: only one install group can be default", c.Name),
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
		switch mode {
		case "none":
			if hasInstallID || hasInstallName || hasLabels {
				return ErrConfig{Description: fmt.Sprintf("branch %q: preview mode none cannot set install_id, install_name, or label_selector", c.Name)}
			}
		case "plan-only", "apply", "build-only":
		default:
			return ErrConfig{Description: fmt.Sprintf("branch %q: unknown preview mode %q (valid modes: none, plan-only, apply, build-only)", c.Name, mode)}
		}
		if mode != "none" && mode != "build-only" && !hasInstallID && !hasInstallName && !hasLabels {
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
