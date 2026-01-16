package temporalzap

import (
	"context"
	"fmt"
	"time"

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
		// depening on the type, only one of String, Integer, Interface will be set
		// we convert all to string to not miss any info in the default case.
		val := fmt.Sprintf("%s %d %v", f.String, f.Integer, f.Interface)
		if f.Interface != nil {
			val = fmt.Sprintf("%v", f.Interface)
		}
		// Use String or Integer based on well known Field types
		switch f.Type {
		case zapcore.StringType:
			val = f.String
		case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type,
			zapcore.Float64Type, zapcore.Float32Type:
			val = fmt.Sprintf("%d", f.Integer)
		case zapcore.BoolType:
			val = fmt.Sprintf("%t", f.Integer == 1)
		case zapcore.DurationType:
			val = fmt.Sprintf("%d seconds", f.Integer/int64(time.Second))
		}
		kvs = append(kvs, f.Key, val)
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
