package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 10s
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field signalType
func (a *Activities) getSandboxSignalConfig(ctx context.Context, signalType string) (*app.SandboxModeSignalConfig, error) {
	var cfg app.SandboxModeSignalConfig
	res := a.db.WithContext(ctx).
		Where(app.SandboxModeSignalConfig{SignalType: signalType, Enabled: true}).
		First(&cfg)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get sandbox-config")
	}

	return &cfg, nil
}
