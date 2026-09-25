package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestMergePreviewConfigInstallOverrideClearsOtherTargets(t *testing.T) {
	installName := "staging"
	installID := "install-1"
	defaults := app.AppBranchPreviewConfig{
		Mode:          app.AppBranchRunPreviewModePlanOnly,
		InstallName:   &installName,
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}},
	}

	resolved := mergePreviewConfig(defaults, &app.AppBranchPreviewOverride{
		InstallID: &installID,
	})

	require.Equal(t, &installID, resolved.InstallID)
	require.Nil(t, resolved.InstallName)
	require.Nil(t, resolved.LabelSelector)
}

func TestMergePreviewConfigModeOverridePreservesTarget(t *testing.T) {
	installID := "install-1"
	mode := app.AppBranchRunPreviewModeApply
	defaults := app.AppBranchPreviewConfig{
		Mode:      app.AppBranchRunPreviewModePlanOnly,
		InstallID: &installID,
	}

	resolved := mergePreviewConfig(defaults, &app.AppBranchPreviewOverride{
		Mode: &mode,
	})

	require.Equal(t, app.AppBranchRunPreviewModeApply, resolved.Mode)
	require.Equal(t, &installID, resolved.InstallID)
}

func TestBranchPreviewConfigOrDefaultUsesPlanOnly(t *testing.T) {
	resolved := branchPreviewConfigOrDefault(&app.AppBranchConfig{})
	require.Equal(t, app.AppBranchRunPreviewModePlanOnly, resolved.Mode)
}

func TestBuildAppBranchRunPreviewAllowsOverrideWhenBranchDisabled(t *testing.T) {
	mode := app.AppBranchRunPreviewModeApply
	none := app.AppBranchRunPreviewModeNone
	_, err := (&Helpers{}).BuildAppBranchRunPreview(
		context.Background(),
		"app-1",
		&app.AppBranchConfig{
			PreviewConfig: &app.AppBranchPreviewConfig{Mode: none},
		},
		&PreviewRunInput{
			Source: app.AppBranchRunPreviewSourceBranch,
			Override: &app.AppBranchPreviewOverride{
				Mode: &mode,
			},
		},
	)
	require.ErrorContains(t, err, "install_id, install_name, or label_selector is required")
}

func TestBuildAppBranchRunPreviewRejectsResolvedNone(t *testing.T) {
	none := app.AppBranchRunPreviewModeNone
	_, err := (&Helpers{}).BuildAppBranchRunPreview(
		context.Background(),
		"app-1",
		&app.AppBranchConfig{},
		&PreviewRunInput{
			Source: app.AppBranchRunPreviewSourceBranch,
			Override: &app.AppBranchPreviewOverride{
				Mode: &none,
			},
		},
	)
	require.ErrorContains(t, err, "preview mode none")
}
