package app

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
)

type AppBranchRunPreviewMode string

const (
	AppBranchRunPreviewModePlanOnly  AppBranchRunPreviewMode = "plan-only"
	AppBranchRunPreviewModeApply     AppBranchRunPreviewMode = "apply"
	AppBranchRunPreviewModeBuildOnly AppBranchRunPreviewMode = "build-only"
	AppBranchRunPreviewModeDisabled  AppBranchRunPreviewMode = "disabled"
)

func (m AppBranchRunPreviewMode) Valid() bool {
	switch m {
	case AppBranchRunPreviewModePlanOnly, AppBranchRunPreviewModeApply, AppBranchRunPreviewModeBuildOnly, AppBranchRunPreviewModeDisabled, "":
		return true
	default:
		return false
	}
}

func (m AppBranchRunPreviewMode) Label() string {
	switch m {
	case AppBranchRunPreviewModeBuildOnly:
		return "build and validate"
	case AppBranchRunPreviewModePlanOnly:
		return "plan-only"
	case AppBranchRunPreviewModeApply:
		return "apply"
	case AppBranchRunPreviewModeDisabled:
		return "disabled"
	default:
		return ""
	}
}

type AppBranchRunPreviewSource string

const (
	AppBranchRunPreviewSourcePR     AppBranchRunPreviewSource = "pr"
	AppBranchRunPreviewSourceCommit AppBranchRunPreviewSource = "commit"
	AppBranchRunPreviewSourceBranch AppBranchRunPreviewSource = "branch"
	AppBranchRunPreviewSourceLocal  AppBranchRunPreviewSource = "local"
)

func (s AppBranchRunPreviewSource) Valid() bool {
	switch s {
	case AppBranchRunPreviewSourcePR, AppBranchRunPreviewSourceCommit, AppBranchRunPreviewSourceBranch, AppBranchRunPreviewSourceLocal, "":
		return true
	default:
		return false
	}
}

type AppBranchPreviewConfig struct {
	Mode AppBranchRunPreviewMode `json:"mode,omitempty"`

	InstallID     *string          `json:"install_id,omitempty"`
	InstallName   *string          `json:"install_name,omitempty"`
	LabelSelector *labels.Selector `json:"label_selector,omitempty"`

	SetStatuses  bool `json:"set_statuses"`
	Comment      bool `json:"comment"`
	IgnoreDrafts bool `json:"ignore_drafts"`
	React        bool `json:"react"`
}

func DefaultAppBranchPreviewConfig() AppBranchPreviewConfig {
	return AppBranchPreviewConfig{
		Mode:         AppBranchRunPreviewModePlanOnly,
		SetStatuses:  true,
		Comment:      true,
		IgnoreDrafts: true,
		React:        true,
	}
}

// UnmarshalJSON defaults ignore_drafts and react to true when omitted so existing
// stored preview configs keep the intended opt-out defaults.
func (c *AppBranchPreviewConfig) UnmarshalJSON(data []byte) error {
	type wire struct {
		Mode          AppBranchRunPreviewMode `json:"mode,omitempty"`
		InstallID     *string                 `json:"install_id,omitempty"`
		InstallName   *string                 `json:"install_name,omitempty"`
		LabelSelector *labels.Selector        `json:"label_selector,omitempty"`
		SetStatuses   bool                    `json:"set_statuses"`
		Comment       bool                    `json:"comment"`
		IgnoreDrafts  *bool                   `json:"ignore_drafts"`
		React         *bool                   `json:"react"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	c.Mode = w.Mode
	c.InstallID = w.InstallID
	c.InstallName = w.InstallName
	c.LabelSelector = w.LabelSelector
	c.SetStatuses = w.SetStatuses
	c.Comment = w.Comment
	c.IgnoreDrafts = true
	if w.IgnoreDrafts != nil {
		c.IgnoreDrafts = *w.IgnoreDrafts
	}
	c.React = true
	if w.React != nil {
		c.React = *w.React
	}
	return nil
}

func (c *AppBranchPreviewConfig) Normalize() {
	if c.Mode == "" {
		c.Mode = AppBranchRunPreviewModePlanOnly
	}
}

func (c *AppBranchPreviewConfig) Validate() error {
	if c == nil {
		return nil
	}
	if !c.Mode.Valid() {
		return fmt.Errorf("preview mode %q is invalid", c.Mode)
	}
	hasInstallID := c.InstallID != nil && *c.InstallID != ""
	hasInstallName := c.InstallName != nil && *c.InstallName != ""
	hasLabels := c.LabelSelector != nil && len(c.LabelSelector.MatchLabels) > 0
	if hasInstallID && hasLabels {
		return fmt.Errorf("preview config: label_selector is mutually exclusive with install_id")
	}
	if hasInstallName && hasLabels {
		return fmt.Errorf("preview config: label_selector is mutually exclusive with install_name")
	}
	if hasInstallID && hasInstallName {
		return fmt.Errorf("preview config: install_id is mutually exclusive with install_name")
	}
	if c.Mode != AppBranchRunPreviewModeBuildOnly && c.Mode != AppBranchRunPreviewModeDisabled && !hasInstallID && !hasInstallName && !hasLabels {
		return fmt.Errorf("preview config: install_id, install_name, or label_selector is required for mode %q", c.Mode)
	}
	return nil
}

type AppBranchPreviewOverride struct {
	Mode      *AppBranchRunPreviewMode `json:"mode,omitempty"`
	InstallID *string                  `json:"install_id,omitempty"`
}
