package app

import "fmt"

type AppBranchRunMode string

const (
	AppBranchRunModeAll         AppBranchRunMode = "all"
	AppBranchRunModeTagPrefix   AppBranchRunMode = "on_tag_prefix"
	AppBranchRunModeGithubLabel AppBranchRunMode = "on_github_label"
	AppBranchRunModeManualOnly  AppBranchRunMode = "manual_only"
)

type AppBranchRunConfig struct {
	Mode        AppBranchRunMode `json:"mode,omitempty"`
	TagPrefix   string           `json:"tag_prefix,omitempty"`
	GithubLabel string           `json:"github_label,omitempty"`
}

func (c *AppBranchRunConfig) Normalize() {
	if c != nil && c.Mode == "" {
		c.Mode = AppBranchRunModeAll
	}
}

func (c AppBranchRunConfig) Validate() error {
	switch c.Mode {
	case "", AppBranchRunModeAll, AppBranchRunModeManualOnly:
		if c.TagPrefix != "" || c.GithubLabel != "" {
			return fmt.Errorf("run mode %q cannot set tag_prefix or github_label", c.Mode)
		}
	case AppBranchRunModeTagPrefix:
		if c.TagPrefix == "" {
			return fmt.Errorf("run mode %q requires tag_prefix", c.Mode)
		}
		if c.GithubLabel != "" {
			return fmt.Errorf("run mode %q cannot set github_label", c.Mode)
		}
	case AppBranchRunModeGithubLabel:
		if c.GithubLabel == "" {
			return fmt.Errorf("run mode %q requires github_label", c.Mode)
		}
		if c.TagPrefix != "" {
			return fmt.Errorf("run mode %q cannot set tag_prefix", c.Mode)
		}
	default:
		return fmt.Errorf("unknown app branch run mode %q", c.Mode)
	}
	return nil
}
