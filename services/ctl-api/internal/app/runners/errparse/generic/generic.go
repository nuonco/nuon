package generic

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const GenericErrorType compositeerrors.Type = "generic"

const (
	maxHeadline = 240
	maxBody     = 8000
)

type GenericError struct {
	Body string `json:"body"`
}

var _ compositeerrors.CompositeError = (*GenericError)(nil)

func (e *GenericError) Error() string {
	if h := headline(e.Body); h != "" {
		return h
	}
	return "Job failed"
}

func (e *GenericError) Type() compositeerrors.Type { return GenericErrorType }
func (e *GenericError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *GenericError) Sections() []compositeerrors.Section {
	if e.Body == "" {
		return nil
	}
	return []compositeerrors.Section{
		compositeerrors.CodeSection("Error output", e.Body),
	}
}

func parseGeneric(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	body := cleanBody(ctx.Raw)
	if body == "" {
		return nil
	}
	return &GenericError{Body: body}
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerGeneric, parseGeneric,
		errparse.AlwaysCandidate(),
	))
}

func cleanBody(raw string) string {
	lines := cleanedLines(raw)
	if len(lines) == 0 {
		return ""
	}
	body := strings.Join(lines, "\n")
	return truncate(body, maxBody)
}

func headline(body string) string {
	lines := strings.Split(body, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "Error:") {
			return truncate(l, maxHeadline)
		}
	}
	for _, l := range lines {
		if l != "" {
			return truncate(l, maxHeadline)
		}
	}
	return ""
}

func cleanedLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		t := strings.TrimSpace(line)
		t = strings.TrimPrefix(t, "│")
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// why: truncate caps s to n runes (not bytes) so it never splits a multi-byte rune
// into invalid UTF-8, appending an ellipsis when it cuts.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
