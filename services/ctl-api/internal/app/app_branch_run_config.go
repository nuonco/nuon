package app

import "fmt"

type AppBranchRunMode string

const (
	AppBranchRunModePush        AppBranchRunMode = "push"
	AppBranchRunModeTagPrefix   AppBranchRunMode = "on_tag"
	AppBranchRunModeGithubLabel AppBranchRunMode = "on_github_label"
	AppBranchRunModeManualOnly  AppBranchRunMode = "manual_only"
)

type AppBranchRunConfig struct {
	Mode        AppBranchRunMode `json:"mode,omitempty"`
	TagPrefix   string           `json:"tag_prefix,omitempty"`
	GithubLabel string           `json:"github_label,omitempty"`
}

func (c *AppBranchRunConfig) Normalize() {
	if c == nil {
		return
	}
	if c.Mode == "" || c.Mode == "all" {
		c.Mode = AppBranchRunModePush
	} else if c.Mode == "on_tag_prefix" {
		c.Mode = AppBranchRunModeTagPrefix
	}
}

func (c AppBranchRunConfig) Validate() error {
	c.Normalize()
	switch c.Mode {
	case AppBranchRunModePush, AppBranchRunModeManualOnly:
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
		return fmt.Errorf("unknown app branch run mode %q (valid modes: push, on_tag, on_github_label, manual_only)", c.Mode)
	}
	return nil
}
