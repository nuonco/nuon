package app

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/labels"
)

type AppBranchRunPreviewMode string

const (
	AppBranchRunPreviewModePlanOnly  AppBranchRunPreviewMode = "plan-only"
	AppBranchRunPreviewModePlanInfra AppBranchRunPreviewMode = "plan-infra"
	AppBranchRunPreviewModeApply     AppBranchRunPreviewMode = "apply"
	AppBranchRunPreviewModeBuildOnly AppBranchRunPreviewMode = "build-only"
)

func (m AppBranchRunPreviewMode) Valid() bool {
	switch m {
	case AppBranchRunPreviewModePlanOnly, AppBranchRunPreviewModePlanInfra, AppBranchRunPreviewModeApply, AppBranchRunPreviewModeBuildOnly, "":
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

	SetStatuses bool `json:"set_statuses"`
	Comment     bool `json:"comment"`
}

func DefaultAppBranchPreviewConfig() AppBranchPreviewConfig {
	return AppBranchPreviewConfig{
		Mode:        AppBranchRunPreviewModePlanOnly,
		SetStatuses: true,
		Comment:     true,
	}
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
	return nil
}

type AppBranchPreviewOverride struct {
	Mode      *AppBranchRunPreviewMode `json:"mode,omitempty"`
	InstallID *string                  `json:"install_id,omitempty"`
}
