package terraform

import (
	"bytes"
	"log/slog"
)

// slogWriter adapts terraform's stdout/stderr streams into an slog.Logger,
// emitting one record per line so terraform's live progress shows up in the
// dashboard log stream the same way the AWS SDK method's step logs do.
type slogWriter struct {
	l   *slog.Logger
	buf bytes.Buffer
}

func (w *slogWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			// Partial line; keep it buffered until the newline arrives.
			w.buf.WriteString(line)
			break
		}
		if trimmed := trimEOL(line); trimmed != "" {
			w.l.Info(trimmed)
		}
	}
	return len(p), nil
}

func trimEOL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
