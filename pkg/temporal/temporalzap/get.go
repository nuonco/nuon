package temporalzap

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logWrapper is a wrapper around log.Logger, with a zap interface. While we emit _all_ logs to zap using the client, we
// use this to consolidate logging with zap from the workflows/activities themselves.
type logCore struct {
	l     log.Logger
	attrs []zap.Field
}

func (l *logCore) Sync() error {
	return nil
}

func (o *logCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if o.Enabled(ent.Level) {
		return ce.AddCore(ent, o)
	}
	return ce
}

// we let the underlying logger decide if the log should be passed on
func (o *logCore) Enabled(level zapcore.Level) bool {
	return true
}

func (o *logCore) With(fields []zapcore.Field) zapcore.Core {
	cloned := o.clone()
	cloned.attrs = append(cloned.attrs, fields...)
	return cloned
}

func (o *logCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	fn := o.convertLevel(ent.Level)

	str := ent.Message
	kvs := make([]interface{}, 0)
	for _, f := range fields {
		kvs = append(kvs, fmt.Sprintf("%v", f))
	}

	fn(str, kvs...)
	return nil
}

func (c *logCore) convertLevel(level zapcore.Level) func(string, ...interface{}) {
	switch level {
	case zapcore.DebugLevel:
		return c.l.Debug
	case zapcore.InfoLevel:
		return c.l.Info
	case zapcore.WarnLevel:
		return c.l.Warn
	case zapcore.ErrorLevel:
		return c.l.Error
	case zapcore.DPanicLevel:
		return c.l.Error
	case zapcore.PanicLevel:
		return c.l.Error
	case zapcore.FatalLevel:
		return c.l.Error
	default:
		return c.l.Info
	}
}

func (o *logCore) clone() *logCore {
	return &logCore{
		l:     o.l,
		attrs: o.attrs,
	}
}

var _ zapcore.Core = (*logCore)(nil)

func NewCore(lg log.Logger) zapcore.Core {
	return &logCore{
		attrs: make([]zapcore.Field, 0),
		l:     lg,
	}
}

func GetWorkflowLogger(ctx workflow.Context) *zap.Logger {
	l := workflow.GetLogger(ctx)
	return zap.New(NewCore(l))
}

func GetActivityLogger(ctx context.Context) *zap.Logger {
	l := activity.GetLogger(ctx)
	return zap.New(NewCore(l))
}
