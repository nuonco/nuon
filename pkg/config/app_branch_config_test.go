package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppBranchConfig_ValidateInstallGroups(t *testing.T) {
	tests := map[string]struct {
		groups  []AppBranchInstallGroupConfig
		wantErr string
	}{
		"no groups is allowed": {
			groups: nil,
		},
		"default only": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "default", Order: 0, Default: true},
			},
		},
		"pinned only group needs neither labels nor default": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "default", Order: 0, Default: true},
				{Name: "manual", Order: 1},
			},
		},
		"labels alongside default": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "default", Order: 0, Default: true},
				{Name: "prod", Order: 1, LabelSelector: map[string]string{"env": "prod"}},
			},
		},
		"no default group": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "prod", Order: 0, LabelSelector: map[string]string{"env": "prod"}},
			},
		},
		"pinned only group without a default": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "manual", Order: 0},
			},
		},
		"two default groups": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "a", Order: 0, Default: true},
				{Name: "b", Order: 1, Default: true},
			},
			wantErr: `branch "main": only one install group can be default`,
		},
		"default group may also carry labels": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "a", Order: 0, Default: true, LabelSelector: map[string]string{"env": "prod"}},
			},
		},
		"duplicate names": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "a", Order: 0, Default: true},
				{Name: "a", Order: 1},
			},
			wantErr: `branch "main": install group names must be unique; "a" is duplicated`,
		},
		"duplicate label selectors": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "a", Order: 0, Default: true, LabelSelector: map[string]string{"env": "prod", "tier": "web"}},
				{Name: "b", Order: 1, LabelSelector: map[string]string{"tier": "web", "env": "prod"}},
			},
			wantErr: `branch "main": install groups "a" and "b" have the same label_selector`,
		},
		"distinct label selectors": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "a", Order: 0, Default: true, LabelSelector: map[string]string{"env": "prod"}},
				{Name: "b", Order: 1, LabelSelector: map[string]string{"env": "staging"}},
			},
		},
		"empty selectors on pinned groups are allowed": {
			groups: []AppBranchInstallGroupConfig{
				{Name: "default", Order: 0, Default: true},
				{Name: "manual-a", Order: 1},
				{Name: "manual-b", Order: 2},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &AppBranchConfig{Name: "main", InstallGroups: tc.groups}
			err := cfg.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
