// Package common provides utilities for v3 UI components.
// The logger sends logs to a file since the TUI should not be interrupted.
package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cctx "github.com/powertoolsdev/mono/pkg/ctx"
)

// Logger is a file-based logger for TUI applications.
// It wraps a zap.Logger and provides both instance methods and context integration.
type Logger struct {
	logger  *zap.Logger
	logPath string
}

// NewLogger creates a new Logger instance that writes to a file.
// This is essential for TUI applications where stdout/stderr cannot be used.
// The log file is created in the system's temp directory.
func NewLogger(name string) (*Logger, error) {
	// Create log file in temp directory
	// os.TempDir() // use this eventually
	logDir := filepath.Join("/tmp", "nuon-cli-logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", name))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Configure encoder for file output
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create core that writes to file
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		zapcore.DebugLevel,
	)

	// Build logger with caller information
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		logger:  zapLogger,
		logPath: logPath,
	}, nil
}

// LogPath returns the path to the log file.
func (l *Logger) LogPath() string {
	return l.logPath
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// WithContext adds this logger to the context.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return cctx.SetLogger(ctx, l.logger)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// --- Backward compatible package-level functions ---

// NewFileLogger creates a new zap logger that writes to a file.
// Deprecated: Use NewLogger instead.
func NewFileLogger(name string) (*zap.Logger, string, error) {
	logger, err := NewLogger(name)
	if err != nil {
		return nil, "", err
	}
	return logger.logger, logger.logPath, nil
}

// SetupLogger creates a new file logger and adds it to the context.
// Returns the updated context and the path to the log file.
// Deprecated: Use NewLogger and WithContext instead.
func SetupLogger(ctx context.Context, name string) (context.Context, string, error) {
	logger, err := NewLogger(name)
	if err != nil {
		return ctx, "", err
	}
	ctx = logger.WithContext(ctx)
	return ctx, logger.logPath, nil
}

// GetLogger retrieves the logger from context.
// If no logger is found, it returns a no-op logger to prevent panics.
func GetLogger(ctx context.Context) *zap.Logger {
	logger, err := cctx.Logger(ctx)
	if err != nil {
		// Return no-op logger if not found in context
		return zap.NewNop()
	}
	return logger
}

// LogInfo logs an info-level message using the logger from context.
func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Info(msg, fields...)
}

// LogDebug logs a debug-level message using the logger from context.
func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Debug(msg, fields...)
}

// LogError logs an error-level message using the logger from context.
func LogError(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Error(msg, fields...)
}

// LogWarn logs a warning-level message using the logger from context.
func LogWarn(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger(ctx).Warn(msg, fields...)
}
