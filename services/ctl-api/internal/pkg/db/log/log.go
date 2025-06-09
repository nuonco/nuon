package log

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func New(l *zap.Logger, cfg *internal.Config) zapgorm2.Logger {
	dl := zapgorm2.New(l)
	dl.IgnoreRecordNotFoundError = true
	// No logging is enabled, unless in debug mode
	dl = dl.LogMode(-1).(zapgorm2.Logger)

	if cfg.LogLevel == "DEBUG" {
		dl = dl.LogMode(logger.Info).(zapgorm2.Logger)
	}
	if cfg.DBLogQueries {
		dl = dl.LogMode(logger.Info).(zapgorm2.Logger)
	}

	// default to 5 seconds
	dl.SlowThreshold = 5 * time.Second
	if cfg.SlowQueryThresholdMS > 0 {
		dl.SlowThreshold = time.Duration(cfg.SlowQueryThresholdMS) * time.Millisecond
	}

	return dl
}
