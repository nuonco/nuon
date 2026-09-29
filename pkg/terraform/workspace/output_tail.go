package workspace

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nuonco/nuon/pkg/terraform/workspace/output"
)

const maxOutputTailBytes = 8 * 1024

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
