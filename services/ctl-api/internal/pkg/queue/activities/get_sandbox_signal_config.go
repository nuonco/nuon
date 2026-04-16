package activities

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 10s
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field signalType
func (a *Activities) getSandboxSignalConfig(ctx context.Context, signalType string) (*app.SandboxSignalConfig, error) {
	var cfg app.SandboxSignalConfig
	res := a.db.WithContext(ctx).
		Where(app.SandboxSignalConfig{SignalType: signalType, Enabled: true}).
		First(&cfg)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}
	return &cfg, nil
}
