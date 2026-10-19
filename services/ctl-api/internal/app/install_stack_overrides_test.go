package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyInstallRunnerImageOverrides(t *testing.T) {
	enabled, disabled := true, false
	pinned, empty := "ami-0install", ""

	tests := map[string]struct {
		appCfg     AppRunnerConfig
		installCfg *InstallConfig
		want       AppRunnerConfig
	}{
		"no install config keeps app values": {
			appCfg: AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0app"},
			want:   AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0app"},
		},
		"unset overrides keep app values": {
			appCfg:     AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0app"},
			installCfg: &InstallConfig{},
			want:       AppRunnerConfig{ImageAutoUpdate: true, ImageID: "ami-0app"},
		},
		"install enables auto-update": {
			installCfg: &InstallConfig{RunnerImageAutoUpdate: &enabled},
			want:       AppRunnerConfig{ImageAutoUpdate: true},
		},
		"install disables app auto-update": {
			appCfg:     AppRunnerConfig{ImageAutoUpdate: true},
			installCfg: &InstallConfig{RunnerImageAutoUpdate: &disabled},
			want:       AppRunnerConfig{ImageAutoUpdate: false},
		},
		"install pin wins over app pin": {
			appCfg:     AppRunnerConfig{ImageID: "ami-0app"},
			installCfg: &InstallConfig{RunnerImageID: &pinned},
			want:       AppRunnerConfig{ImageID: "ami-0install"},
		},
		"empty install pin keeps app pin": {
			appCfg:     AppRunnerConfig{ImageID: "ami-0app"},
			installCfg: &InstallConfig{RunnerImageID: &empty},
			want:       AppRunnerConfig{ImageID: "ami-0app"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			runnerCfg := tc.appCfg

			ApplyInstallRunnerImageOverrides(&Install{InstallConfig: tc.installCfg}, &runnerCfg)

			assert.Equal(t, tc.want.ImageAutoUpdate, runnerCfg.ImageAutoUpdate)
			assert.Equal(t, tc.want.ImageID, runnerCfg.ImageID)
		})
	}
}
