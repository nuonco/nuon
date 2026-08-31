package log

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func TestNewFXLogUsesConfiguredLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		debugEnabled bool
	}{
		{name: "production", logLevel: "INFO", debugEnabled: false},
		{name: "development", logLevel: "DEBUG", debugEnabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &internal.Config{}
			cfg.LogLevel = tt.logLevel
			logger, err := NewFXLog(cfg)
			require.NoError(t, err)

			zapLogger, ok := logger.(*fxevent.ZapLogger)
			require.True(t, ok)
			require.Equal(t, tt.debugEnabled, zapLogger.Logger.Core().Enabled(zap.DebugLevel))
			require.True(t, zapLogger.Logger.Core().Enabled(zap.ErrorLevel))
		})
	}
}
