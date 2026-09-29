package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cctx "github.com/nuonco/nuon/pkg/ctx"
)

type Logger struct {
	logger  *zap.Logger
	logPath string
}

func NewLogger(name string) (*Logger, error) {
	logDir := filepath.Join("/tmp", "nuon-cli-logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", name))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		zapcore.DebugLevel,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		logger:  zapLogger,
		logPath: logPath,
	}, nil
}

func (l *Logger) LogPath() string {
	return l.logPath
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return cctx.SetLogger(ctx, l.logger)
}

func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Deprecated: Use NewLogger instead.
func NewFileLogger(name string) (*zap.Logger, string, error) {
	logger, err := NewLogger(name)
	if err != nil {
		return nil, "", err
	}
	return logger.logger, logger.logPath, nil
}

// Deprecated: Use NewLogger and WithContext instead.
func SetupLogger(ctx context.Context, name string) (context.Context, string, error) {
	logger, err := NewLogger(name)
	if err != nil {
		return ctx, "", err
	}
	ctx = logger.WithContext(ctx)
	return ctx, logger.logPath, nil
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, err := cctx.Logger(ctx)
	if err != nil {
		return zap.NewNop()
	}
	return logger
}

func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Info(msg, fields...)
}

func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Debug(msg, fields...)
}

func LogError(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Error(msg, fields...)
}

func LogWarn(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Warn(msg, fields...)
}
