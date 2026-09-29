package errcapture

import (
	"context"
	"strings"
	"sync"
	"unicode/utf8"

	"go.uber.org/zap/zapcore"
)

// why: MetadataKey is the error-metadata key the captured output is attached under.
// It must match the key ctl-api prefers when parsing a failed result
// (services/ctl-api/.../create_runner_job_execution_result.go: errMetaKeyOutput).
const MetadataKey = "error_output"

const defaultMaxBytes = 64 * 1024

type Capture struct {
	mu   sync.Mutex
	buf  []string
	size int
	max  int
	full bool
}

func New() *Capture {
	return &Capture{max: defaultMaxBytes}
}

func (c *Capture) Core() zapcore.Core {
	return &captureCore{LevelEnabler: zapcore.ErrorLevel, cap: c}
}

func (c *Capture) String() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.buf, "\n")
}

// why: append records a line, respecting the size bound. When a line would overflow
// the bound, a UTF-8-safe prefix that fits is kept rather than dropping the
// line whole — otherwise a single oversized diagnostic (the root cause) could
// be lost entirely, leaving nothing for ctl-api to parse. After an overflow no
// further lines are accepted (the head is preserved).
func (c *Capture) append(line string) {
	if line == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.full {
		return
	}
	if c.size+len(line)+1 > c.max {
		budget := c.max - c.size - 1
		if prefix := safeUTF8Prefix(line, budget); prefix != "" {
			c.buf = append(c.buf, prefix)
			c.size += len(prefix) + 1
		}
		c.full = true
		return
	}
	c.buf = append(c.buf, line)
	c.size += len(line) + 1
}

func safeUTF8Prefix(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	b := maxBytes
	for b > 0 && !utf8.RuneStart(s[b]) {
		b--
	}
	return s[:b]
}

type ctxKey struct{}

func NewContext(ctx context.Context, cap *Capture) context.Context {
	return context.WithValue(ctx, ctxKey{}, cap)
}

func FromContext(ctx context.Context) *Capture {
	cap, _ := ctx.Value(ctxKey{}).(*Capture)
	return cap
}

func Output(ctx context.Context) string {
	return FromContext(ctx).String()
}

type captureCore struct {
	zapcore.LevelEnabler
	cap    *Capture
	fields []zapcore.Field
}

func (c *captureCore) With(fs []zapcore.Field) zapcore.Core {
	nf := make([]zapcore.Field, 0, len(c.fields)+len(fs))
	nf = append(nf, c.fields...)
	nf = append(nf, fs...)
	return &captureCore{LevelEnabler: c.LevelEnabler, cap: c.cap, fields: nf}
}

func (c *captureCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *captureCore) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	line := ent.Message
	if e := errorField(fs, c.fields); e != "" && e != line {
		if line == "" {
			line = e
		} else {
			line = line + ": " + e
		}
	}
	line = diagnosticField(fs, c.fields).render(line)
	c.cap.append(line)
	return nil
}

func (c *captureCore) Sync() error { return nil }

func errorField(groups ...[]zapcore.Field) string {
	for _, g := range groups {
		for _, f := range g {
			if f.Key != "error" {
				continue
			}
			if err, ok := f.Interface.(error); ok && err != nil {
				return err.Error()
			}
			if f.String != "" {
				return f.String
			}
		}
	}
	return ""
}

const diagnosticKey = "diagnostic"

type diagnostic struct {
	summary string
	detail  string
	address string
}

func diagnosticField(groups ...[]zapcore.Field) diagnostic {
	for _, g := range groups {
		for _, f := range g {
			if f.Key != diagnosticKey {
				continue
			}
			m, ok := f.Interface.(map[string]interface{})
			if !ok {
				continue
			}
			return diagnostic{
				summary: diagnosticString(m, "summary"),
				detail:  diagnosticString(m, "detail"),
				address: diagnosticString(m, "address"),
			}
		}
	}
	return diagnostic{}
}

func diagnosticString(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return strings.TrimSpace(s)
}

func (d diagnostic) render(msg string) string {
	lines := make([]string, 0, 3)
	if msg != "" {
		lines = append(lines, msg)
	}
	if d.summary != "" && !strings.Contains(msg, d.summary) {
		lines = append(lines, errorPrefixed(d.summary))
	}
	if d.address != "" {
		lines = append(lines, "  with "+d.address)
	}
	if d.detail != "" {
		lines = append(lines, d.detail)
	}
	return strings.Join(lines, "\n")
}

func errorPrefixed(summary string) string {
	if strings.HasPrefix(summary, "Error:") {
		return summary
	}
	return "Error: " + summary
}
