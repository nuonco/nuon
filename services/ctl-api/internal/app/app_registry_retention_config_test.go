package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
)

func TestResolveRegistryRetention(t *testing.T) {
	str := func(s string) *string { return &s }
	num := func(n int) *int { return &n }

	tests := map[string]struct {
		install    *InstallConfig
		app        *AppRegistryRetentionConfig
		component  *ComponentConfigConnection
		wantMode   string
		wantRetain int
	}{
		"defaults when nothing is set": {
			wantMode:   config.DefaultRegistryRetentionMode,
			wantRetain: config.DefaultRegistryRetentionRetainCount,
		},
		"app section sets both": {
			app:        &AppRegistryRetentionConfig{Mode: str(config.RegistryRetentionModeDryRun), RetainCount: num(5)},
			wantMode:   config.RegistryRetentionModeDryRun,
			wantRetain: 5,
		},
		"app section with only mode keeps default retain count": {
			app:        &AppRegistryRetentionConfig{Mode: str(config.RegistryRetentionModeEnforce)},
			wantMode:   config.RegistryRetentionModeEnforce,
			wantRetain: config.DefaultRegistryRetentionRetainCount,
		},
		"install mode overrides app mode": {
			install:    &InstallConfig{RegistryRetentionMode: str(config.RegistryRetentionModeEnforce)},
			app:        &AppRegistryRetentionConfig{Mode: str(config.RegistryRetentionModeDryRun), RetainCount: num(5)},
			wantMode:   config.RegistryRetentionModeEnforce,
			wantRetain: 5,
		},
		"install explicit off beats app enforce": {
			install:    &InstallConfig{RegistryRetentionMode: str(config.RegistryRetentionModeOff)},
			app:        &AppRegistryRetentionConfig{Mode: str(config.RegistryRetentionModeEnforce)},
			wantMode:   config.RegistryRetentionModeOff,
			wantRetain: config.DefaultRegistryRetentionRetainCount,
		},
		"install override without app section": {
			install:    &InstallConfig{RegistryRetentionMode: str(config.RegistryRetentionModeDryRun)},
			wantMode:   config.RegistryRetentionModeDryRun,
			wantRetain: config.DefaultRegistryRetentionRetainCount,
		},
		"install config without override inherits app": {
			install:    &InstallConfig{},
			app:        &AppRegistryRetentionConfig{Mode: str(config.RegistryRetentionModeEnforce)},
			wantMode:   config.RegistryRetentionModeEnforce,
			wantRetain: config.DefaultRegistryRetentionRetainCount,
		},
		"component retain count overrides app": {
			app:        &AppRegistryRetentionConfig{RetainCount: num(2)},
			component:  &ComponentConfigConnection{RegistryRetainCount: num(10)},
			wantMode:   config.DefaultRegistryRetentionMode,
			wantRetain: 10,
		},
		"component without override inherits app retain count": {
			app:        &AppRegistryRetentionConfig{RetainCount: num(2)},
			component:  &ComponentConfigConnection{},
			wantMode:   config.DefaultRegistryRetentionMode,
			wantRetain: 2,
		},
		"component override does not touch mode": {
			install:    &InstallConfig{RegistryRetentionMode: str(config.RegistryRetentionModeEnforce)},
			component:  &ComponentConfigConnection{RegistryRetainCount: num(1)},
			wantMode:   config.RegistryRetentionModeEnforce,
			wantRetain: 1,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mode, retain := ResolveRegistryRetention(tc.install, tc.app, tc.component)
			assert.Equal(t, tc.wantMode, mode)
			assert.Equal(t, tc.wantRetain, retain)
		})
	}
}

func TestResolveRegistryGracePeriod(t *testing.T) {
	num := func(n int) *int { return &n }
	day := 24 * time.Hour

	tests := map[string]struct {
		app  *AppRegistryRetentionConfig
		want time.Duration
	}{
		"default when no app section": {
			want: config.DefaultRegistryRetentionGracePeriodDays * day,
		},
		"default when app section omits it": {
			app:  &AppRegistryRetentionConfig{RetainCount: num(5)},
			want: config.DefaultRegistryRetentionGracePeriodDays * day,
		},
		"app sets days": {
			app:  &AppRegistryRetentionConfig{GracePeriodDays: num(7)},
			want: 7 * day,
		},
		"zero disables the grace period": {
			app:  &AppRegistryRetentionConfig{GracePeriodDays: num(0)},
			want: 0,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, ResolveRegistryGracePeriod(tc.app))
		})
	}
}
