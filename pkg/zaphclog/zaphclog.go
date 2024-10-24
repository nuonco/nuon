package zaphclog

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
)

// Wrap simplifies wrapping zap instance to hclog.Logger interfacs.
func Wrap(zap *zap.Logger) hclog.Logger {
	return Wrapper{Zap: zap}
}

type Level = hclog.Level

// Wrapper holds *zap.Logger and adapts its methods to declared by hclog.Logger.
type Wrapper struct {
	Zap  *zap.Logger
	name string
}

func (w Wrapper) Debug(msg string, args ...interface{}) {
	w.Zap.Debug(msg, convertToZapAny(args...)...)
}

func (w Wrapper) Info(msg string, args ...interface{}) {
	w.Zap.Info(msg, convertToZapAny(args...)...)
}

func (w Wrapper) Warn(msg string, args ...interface{}) {
	w.Zap.Warn(msg, convertToZapAny(args...)...)
}

func (w Wrapper) Error(msg string, args ...interface{}) {
	w.Zap.Error(msg, convertToZapAny(args...)...)
}

// Log logs messages with four simplified levels - Debug,Warn,Error and Info as a default.
func (w Wrapper) Log(lvl Level, msg string, args ...interface{}) {
	switch lvl {
	case hclog.Debug:
		w.Debug(msg, args...)
	case hclog.Warn:
		w.Warn(msg, args...)
	case hclog.Error:
		w.Error(msg, args...)
	case hclog.DefaultLevel, hclog.Info, hclog.NoLevel, hclog.Off, hclog.Trace:
		w.Info(msg, args...)
	}
}

// Trace will log an info-level message in Zap.
func (w Wrapper) Trace(msg string, args ...interface{}) {
	w.Zap.Info(msg, convertToZapAny(args...)...)
}

// With returns a logger with always-presented key-value pairs.
func (w Wrapper) With(args ...interface{}) hclog.Logger {
	return &Wrapper{Zap: w.Zap.With(convertToZapAny(args...)...)}
}

// Named returns a logger with the specific nams.
// The name string will always be presented in a log messages.
func (w Wrapper) Named(name string) hclog.Logger {
	return &Wrapper{Zap: w.Zap.Named(name), name: name}
}

// Name returns a logger's name (if presented).
func (w Wrapper) Name() string { return w.name }

// ResetNamed has the same implementation as Named.
func (w Wrapper) ResetNamed(name string) hclog.Logger {
	return &Wrapper{Zap: w.Zap.Named(name), name: name}
}

// StandardWriter returns os.Stderr as io.Writer.
func (w Wrapper) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return &zaphclogWriter{
		l: w,
	}
}

// StandardLogger returns standard logger with os.Stderr as a writer.
func (w Wrapper) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(w.StandardWriter(opts), "", log.LstdFlags)
}

// IsTrace has no implementation.
func (w Wrapper) IsTrace() bool { return false }

// IsDebug has no implementation.
func (w Wrapper) IsDebug() bool { return false }

// IsInfo has no implementation.
func (w Wrapper) IsInfo() bool { return false }

// IsWarn has no implementation.
func (w Wrapper) IsWarn() bool { return false }

// IsError has no implementation.
func (w Wrapper) IsError() bool { return false }

// ImpliedArgs has no implementation.
func (w Wrapper) ImpliedArgs() []interface{} { return nil }

// SetLevel has no implementation.
func (w Wrapper) SetLevel(lvl Level) {
}

func (w Wrapper) GetLevel() hclog.Level {
	return hclog.LevelFromString(w.Zap.Level().String())
}
