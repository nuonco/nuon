package zapwriter

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ io.Writer = (*zapWriter)(nil)

func New(l *zap.Logger, level zapcore.Level, prefix string) *zapWriter {
	return NewWithOpts(l,
		WithLogLevel(level),
		WithPrefix(prefix),
	)
}

func NewWithOpts(l *zap.Logger, opts ...optFn) *zapWriter {
	zw := &zapWriter{
		l:     l,
		level: zapcore.InfoLevel,
	}

	for _, opt := range opts {
		opt(zw)
	}

	return zw
}

type zapWriter struct {
	l      *zap.Logger
	level  zapcore.Level
	prefix string

	lineFormatter func(string) string
	lineLeveler   func(string) zapcore.Level
}

type optFn func(*zapWriter)

func WithPrefix(prefix string) optFn {
	return func(r *zapWriter) {
		r.lineFormatter = func(str string) string {
			return prefix + str
		}
	}
}

func WithLogLevel(level zapcore.Level) optFn {
	return func(r *zapWriter) {
		r.level = level
	}
}

func WithLineFormatter(fn func(string) string) optFn {
	return func(r *zapWriter) {
		r.lineFormatter = fn
	}
}

func WithLineLeveler(fn func(string) zapcore.Level) optFn {
	return func(r *zapWriter) {
		r.lineLeveler = fn
	}
}
