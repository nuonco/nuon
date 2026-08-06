package log

import (
	"os"
	"path/filepath"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pkg/errors"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
)

func NewOTELJobLogger(cfg *runnerconfig.Config, lp *log.LoggerProvider, jobID string) (*zap.Logger, error) {
	zapCore := otelzap.NewCore("oteljob",
		otelzap.WithLoggerProvider(lp))

	core := zapcore.Core(zapCore)

	if cfg.JobLogDir != "" && jobID != "" {
		fileCore, err := newJobFileCore(cfg.JobLogDir, jobID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open job log file")
		}
		core = zapcore.NewTee(core, fileCore)
	}

	dev, err := NewDev(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dev logger")
	}

	// if running inside of nuonctl, we automatically print all logs to stdout as well
	if cfg.IsNuonctl {
		core = zapcore.NewTee(core, dev.Core())
	}

	return zap.New(core), nil
}

func newJobFileCore(dir, jobID string) (zapcore.Core, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, jobID+".ndjson"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	return zapcore.NewCore(enc, zapcore.Lock(f), zapcore.DebugLevel), nil
}
