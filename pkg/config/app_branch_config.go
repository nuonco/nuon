package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type AppBranchInstallGroupConfig struct {
	Name           string `mapstructure:"name" toml:"name" jsonschema:"required"`
	Order          int    `mapstructure:"order" toml:"order"`
	UseForPreviews bool   `mapstructure:"use_for_previews,omitempty" toml:"use_for_previews,omitempty"`

	InstallIDs    []string          `mapstructure:"install_ids,omitempty" toml:"install_ids,omitempty"`
	InstallNames  []string          `mapstructure:"install_names,omitempty" toml:"install_names,omitempty"`
	LabelSelector map[string]string `mapstructure:"label_selector,omitempty" toml:"label_selector,omitempty"`
}

func (c AppBranchInstallGroupConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install group")
	addDescription(schema, "order", "deployment order (lower runs first)")
	addDescription(schema, "use_for_previews", "use this group for plan-only preview runs")
	addDescription(schema, "install_ids", "static list of install IDs")
	addDescription(schema, "install_names", "static list of install names, resolved to IDs at sync time")
	addDescription(schema, "label_selector", "label key-value pairs to dynamically match installs")
}

type AppBranchConfig struct {
	Name          string               `mapstructure:"name" toml:"name" jsonschema:"required"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`

	InstallGroups []AppBranchInstallGroupConfig `mapstructure:"install_groups,omitempty" toml:"install_groups,omitempty"`
}

func (c AppBranchConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the app branch")
	addDescription(schema, "connected_repo", "connected GitHub repo the branch tracks")
	addDescription(schema, "public_repo", "public git repo the branch tracks")
	addDescription(schema, "install_groups", "ordered deployment groups for this branch")
}

func (c *AppBranchConfig) Validate() error {
	for _, g := range c.InstallGroups {
		hasStatic := len(g.InstallIDs) > 0 || len(g.InstallNames) > 0
		hasLabels := len(g.LabelSelector) > 0
		if hasStatic && hasLabels {
			return ErrConfig{
				Description: fmt.Sprintf("install group %q: label_selector is mutually exclusive with install_ids and install_names", g.Name),
			}
		}
	}
	return nil
}

func (c *AppBranchConfig) parse() error {
	return nil
}
