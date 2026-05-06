package helm

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapSlogHandler is a small slog.Handler that forwards every record to a
// *zap.Logger so that the Helm v4 SDK's internal slog output flows through the
// runner's existing OTEL pipeline instead of going to os.Stdout.
type zapSlogHandler struct {
	zl    *zap.Logger
	attrs []zapcore.Field
	group string
}

// newZapSlogHandler returns a slog.Handler bridging to the given *zap.Logger.
func newZapSlogHandler(zl *zap.Logger) slog.Handler {
	return &zapSlogHandler{zl: zl}
}

// Enabled reports whether the handler should process records at the given
// level. We accept everything and let zap's level configuration filter.
func (h *zapSlogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle converts an slog.Record into a zap log entry at the equivalent level.
func (h *zapSlogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make([]zapcore.Field, 0, len(h.attrs)+r.NumAttrs())
	fields = append(fields, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		fields = append(fields, slogAttrToZapField(h.group, a))
		return true
	})

	switch {
	case r.Level >= slog.LevelError:
		h.zl.Error(r.Message, fields...)
	case r.Level >= slog.LevelWarn:
		h.zl.Warn(r.Message, fields...)
	case r.Level >= slog.LevelInfo:
		h.zl.Info(r.Message, fields...)
	default:
		h.zl.Debug(r.Message, fields...)
	}
	return nil
}

// WithAttrs returns a new handler whose records always carry the given attrs.
func (h *zapSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := &zapSlogHandler{
		zl:    h.zl,
		group: h.group,
		attrs: append([]zapcore.Field(nil), h.attrs...),
	}
	for _, a := range attrs {
		out.attrs = append(out.attrs, slogAttrToZapField(h.group, a))
	}
	return out
}

// WithGroup namespaces subsequent attribute keys under the given group name.
func (h *zapSlogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	g := name
	if h.group != "" {
		g = h.group + "." + name
	}
	return &zapSlogHandler{
		zl:    h.zl,
		group: g,
		attrs: append([]zapcore.Field(nil), h.attrs...),
	}
}

func slogAttrToZapField(group string, a slog.Attr) zapcore.Field {
	key := a.Key
	if group != "" {
		key = group + "." + key
	}
	return zap.Any(key, a.Value.Any())
}
