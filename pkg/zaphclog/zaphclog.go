package zaphclog

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
)

func Wrap(zap *zap.Logger) hclog.Logger {
	return Wrapper{Zap: zap}
}

type Level = hclog.Level

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

func (w Wrapper) Trace(msg string, args ...interface{}) {
	w.Zap.Info(msg, convertToZapAny(args...)...)
}

func (w Wrapper) With(args ...interface{}) hclog.Logger {
	return &Wrapper{Zap: w.Zap.With(convertToZapAny(args...)...)}
}

func (w Wrapper) Named(name string) hclog.Logger {
	return &Wrapper{Zap: w.Zap.Named(name), name: name}
}

func (w Wrapper) Name() string { return w.name }

func (w Wrapper) ResetNamed(name string) hclog.Logger {
	return &Wrapper{Zap: w.Zap.Named(name), name: name}
}

func (w Wrapper) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return &zaphclogWriter{
		l:  w,
		zl: w.Zap,
	}
}

func (w Wrapper) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(w.StandardWriter(opts), "", log.LstdFlags)
}

func (w Wrapper) IsTrace() bool { return false }

func (w Wrapper) IsDebug() bool { return false }

func (w Wrapper) IsInfo() bool { return false }

func (w Wrapper) IsWarn() bool { return false }

func (w Wrapper) IsError() bool { return false }

func (w Wrapper) ImpliedArgs() []interface{} { return nil }

func (w Wrapper) SetLevel(lvl Level) {
}

func (w Wrapper) GetLevel() hclog.Level {
	return hclog.LevelFromString(w.Zap.Level().String())
}
