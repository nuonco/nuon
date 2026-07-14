// Package errcapture captures a runner job execution's error-level log output
// so it can be attached to the failed execution result as rich diagnostics.
//
// Why this exists: tools like terraform emit their real error detail (e.g. an
// AWS "AccessDenied ... is not authorized to perform: s3:CreateBucket") into
// the log stream via structured @level:"error" records, while the Go error the
// runner wraps up is often just "exit status 1". ctl-api parses the failed
// result's error text into a structured composite error, so it needs the rich
// text, not the thin wrapper.
//
// A Capture is a zapcore.Core teed into the per-execution job logger. It
// records error-level entries into a bounded in-memory buffer. The runner's API
// client decorator reads the buffer and attaches it to the result under
// MetadataKey when a job reports failure — one universal chokepoint, no
// per-handler wiring.
package errcapture

import (
	"context"
	"strings"
	"sync"
	"unicode/utf8"

	"go.uber.org/zap/zapcore"
)

// MetadataKey is the error-metadata key the captured output is attached under.
// It must match the key ctl-api prefers when parsing a failed result
// (services/ctl-api/.../create_runner_job_execution_result.go: errMetaKeyOutput).
const MetadataKey = "error_output"

// defaultMaxBytes bounds the captured output so a pathological log can't grow
// unbounded in memory or bloat the result payload. The buffer keeps the head
// (the first errors, usually the root cause) and drops the rest once full.
const defaultMaxBytes = 64 * 1024

// Capture accumulates error-level log lines for one job execution. It is safe
// for concurrent use: the job logger it feeds is shared across the execution's
// goroutines (steps, monitor).
type Capture struct {
	mu   sync.Mutex
	buf  []string
	size int
	max  int
	full bool
}

// New returns an empty Capture with the default size bound.
func New() *Capture {
	return &Capture{max: defaultMaxBytes}
}

// Core returns the zapcore.Core to tee into the job logger. It only accepts
// error-level and above entries.
func (c *Capture) Core() zapcore.Core {
	return &captureCore{LevelEnabler: zapcore.ErrorLevel, cap: c}
}

// String returns the captured lines joined by newlines. Safe to call on a nil
// Capture (returns "").
func (c *Capture) String() string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.buf, "\n")
}

// append records a line, respecting the size bound. When a line would overflow
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
		budget := c.max - c.size - 1 // reserve one byte for the join newline
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

// safeUTF8Prefix returns the longest prefix of s that is at most maxBytes long
// and does not split a multi-byte rune. Returns "" when maxBytes <= 0.
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

// NewContext returns ctx carrying cap so downstream code (the API client
// decorator) can read the captured output.
func NewContext(ctx context.Context, cap *Capture) context.Context {
	return context.WithValue(ctx, ctxKey{}, cap)
}

// FromContext returns the Capture on ctx, or nil when none is set.
func FromContext(ctx context.Context) *Capture {
	cap, _ := ctx.Value(ctxKey{}).(*Capture)
	return cap
}

// Output is a nil-safe shortcut for FromContext(ctx).String().
func Output(ctx context.Context) string {
	return FromContext(ctx).String()
}

// captureCore is a zapcore.Core that appends error-level entries to a Capture.
// It carries accumulated With() fields so it can surface the value of a zap
// "error" field (e.g. zap.Error(err)) alongside the entry message.
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
	c.cap.append(line)
	return nil
}

func (c *captureCore) Sync() error { return nil }

// errorField returns the rendered value of a zap "error" field, preferring
// entry-level fields over accumulated With() fields. Returns "" when absent.
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
