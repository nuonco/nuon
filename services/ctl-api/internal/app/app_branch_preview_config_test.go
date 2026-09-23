package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
)

func TestAppBranchPreviewConfigUnmarshalDefaults(t *testing.T) {
	var cfg AppBranchPreviewConfig
	require.NoError(t, json.Unmarshal([]byte(`{"comment":true}`), &cfg))
	require.True(t, cfg.Comment)
	require.True(t, cfg.IgnoreDrafts)
	require.True(t, cfg.React)

	require.NoError(t, json.Unmarshal([]byte(`{"ignore_drafts":false,"react":false}`), &cfg))
	require.False(t, cfg.IgnoreDrafts)
	require.False(t, cfg.React)
}

func TestAppBranchPreviewConfigValidateTargets(t *testing.T) {
	installID := "install-1"
	installName := "staging"
	selector := &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}

	for _, tc := range []struct {
		name    string
		cfg     AppBranchPreviewConfig
		wantErr string
	}{
		{
			name:    "plan-only requires target",
			cfg:     AppBranchPreviewConfig{Mode: AppBranchRunPreviewModePlanOnly},
			wantErr: "install_id, install_name, or label_selector is required",
		},
		{
			name:    "apply requires target",
			cfg:     AppBranchPreviewConfig{Mode: AppBranchRunPreviewModeApply},
			wantErr: "install_id, install_name, or label_selector is required",
		},
		{
			name: "build-only does not require target",
			cfg:  AppBranchPreviewConfig{Mode: AppBranchRunPreviewModeBuildOnly},
		},
		{
			name: "none does not require target",
			cfg:  AppBranchPreviewConfig{Mode: AppBranchRunPreviewModeNone},
		},
		{
			name:    "none rejects target",
			cfg:     AppBranchPreviewConfig{Mode: AppBranchRunPreviewModeNone, InstallName: &installName},
			wantErr: "mode none cannot set",
		},
		{
			name: "install ID is valid",
			cfg:  AppBranchPreviewConfig{Mode: AppBranchRunPreviewModePlanOnly, InstallID: &installID},
		},
		{
			name: "install name is valid",
			cfg:  AppBranchPreviewConfig{Mode: AppBranchRunPreviewModePlanOnly, InstallName: &installName},
		},
		{
			name: "label selector is valid",
			cfg:  AppBranchPreviewConfig{Mode: AppBranchRunPreviewModePlanOnly, LabelSelector: selector},
		},
		{
			name: "targets are mutually exclusive",
			cfg: AppBranchPreviewConfig{
				Mode:          AppBranchRunPreviewModePlanOnly,
				InstallID:     &installID,
				LabelSelector: selector,
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "install ID and name are mutually exclusive",
			cfg: AppBranchPreviewConfig{
				Mode:        AppBranchRunPreviewModePlanOnly,
				InstallID:   &installID,
				InstallName: &installName,
			},
			wantErr: "mutually exclusive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
