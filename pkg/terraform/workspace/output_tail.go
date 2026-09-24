package workspace

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nuonco/nuon/pkg/terraform/workspace/output"
)

// maxOutputTailBytes bounds how much of terraform's JSON stdout is folded into a
// failed run's error. The diagnostics terraform emits before exiting are at the
// end of the stream, so the tail is the part worth keeping.
const maxOutputTailBytes = 8 * 1024

// errWithOutputTail wraps err with the tail of terraform's JSON output stream.
// tfexec reports a failed run as a bare "exit status 1", so without the stream
// the error that reaches ctl-api carries nothing to parse a real cause from.
func errWithOutputTail(op string, err error, out output.Dual) error {
	tail := outputTail(out)
	if tail == "" {
		return fmt.Errorf("error running %s: %w", op, err)
	}
	return fmt.Errorf("error running %s: %w\n%s", op, err, tail)
}

func outputTail(out output.Dual) string {
	if out == nil {
		return ""
	}
	byts, err := out.Bytes()
	if err != nil || len(byts) == 0 {
		return ""
	}
	return tailLines(string(byts), maxOutputTailBytes)
}

// tailLines returns the trailing lines of s that fit within maxBytes. It cuts on
// a line boundary where it can, so every retained record stays a parseable JSON
// object; when a single line is larger than the bound it falls back to a
// rune-aligned cut rather than emitting invalid UTF-8.
func tailLines(s string, maxBytes int) string {
	s = strings.TrimRight(s, "\n")
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return strings.TrimSpace(s)
	}

	cut := s[len(s)-maxBytes:]
	if i := strings.IndexByte(cut, '\n'); i >= 0 {
		return strings.TrimSpace(cut[i+1:])
	}
	for len(cut) > 0 && !utf8.RuneStart(cut[0]) {
		cut = cut[1:]
	}
	return strings.TrimSpace(cut)
}
