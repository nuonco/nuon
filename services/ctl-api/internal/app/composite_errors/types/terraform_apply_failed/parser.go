package terraform_apply_failed

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

const (
	parserName    = "terraform.generic"
	parserVersion = "1"
)

// Parser implements the broad fallback for terraform output. It registers
// against the entire "terraform" subtree so it runs for plan, apply, init,
// and destroy contexts.
type Parser struct{}

var _ composite_error.Parser = (*Parser)(nil)

func (Parser) Name() string    { return parserName }
func (Parser) Version() string { return parserVersion }
func (Parser) Contexts() []composite_error.ParseContext {
	return []composite_error.ParseContext{"terraform"}
}

func (p Parser) Parse(_ context.Context, in composite_error.ParseInput) composite_error.ParseResult {
	raw := string(in.Raw)
	if !looksLikeTerraform(raw) {
		return composite_error.ParseResult{Matched: false}
	}

	stage := stageFromContext(in.Invocation)
	diagnostics := extractDiagnostics(raw)

	e := &Error{
		Stage:       stage,
		Diagnostics: diagnostics,
	}
	if len(diagnostics) == 0 {
		e.Message = firstNonEmptyLine(raw)
	}

	return composite_error.ParseResult{
		Matched: true,
		Error:   e,
		Source: composite_error.Source{
			ParserName:    parserName,
			ParserVersion: parserVersion,
			Snippet:       composite_error.CapSnippet(raw),
		},
	}
}

// looksLikeTerraform detects terraform CLI output. We accept either the
// box-drawing diagnostic markers (╷ │ ╵) or one of the standard Error/Warning
// prefixes followed by a colon.
func looksLikeTerraform(s string) bool {
	if strings.ContainsRune(s, '╷') || strings.ContainsRune(s, '╵') {
		return true
	}
	for _, marker := range []string{"\nError:", "\n│ Error:", "Error: "} {
		if strings.Contains(s, marker) {
			return true
		}
	}
	// Also accept first-line Error: prefix.
	if strings.HasPrefix(strings.TrimSpace(s), "Error:") {
		return true
	}
	return false
}

func stageFromContext(inv composite_error.InvocationContext) string {
	if inv.Extra != nil {
		if v, ok := inv.Extra["terraform_stage"].(string); ok && v != "" {
			return v
		}
	}
	// Best-effort guess from the OwnerType — refined as we wire more
	// invocation sites.
	switch {
	case strings.Contains(inv.OwnerType, "plan"):
		return "plan"
	case strings.Contains(inv.OwnerType, "apply"):
		return "apply"
	case strings.Contains(inv.OwnerType, "destroy"):
		return "destroy"
	}
	return ""
}

// firstNonEmptyLine returns the first non-empty line of s, trimmed.
func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// extractDiagnostics walks s and returns one Diagnostic per ╷…╵ block.
//
// Each block looks like:
//
//	╷
//	│ Error: <summary>
//	│
//	│   with <resource>,
//	│   on <file> line <n>, in resource "..." "...":
//	│  <n>: ...
//	│
//	╵
//
// Lines inside the block start with "│ " (or "│" with no space).
// We strip the prefix and parse "Error:", "with ...,", and "on ... line N,".
func extractDiagnostics(s string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(s, "\n")

	inBlock := false
	var current []string
	for _, line := range lines {
		trim := strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(strings.TrimSpace(trim), "╷"):
			inBlock = true
			current = nil
		case strings.HasPrefix(strings.TrimSpace(trim), "╵"):
			if inBlock {
				if d, ok := parseBlock(current); ok {
					diags = append(diags, d)
				}
			}
			inBlock = false
			current = nil
		case inBlock:
			current = append(current, stripBlockPrefix(trim))
		}
	}

	// Also catch ungroup'd "Error: ..." lines that didn't appear inside a
	// box (older terraform versions, certain providers).
	if len(diags) == 0 {
		if d, ok := parseLooseError(s); ok {
			diags = append(diags, d)
		}
	}

	return diags
}

func stripBlockPrefix(line string) string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "│")
	if strings.HasPrefix(t, " ") {
		t = t[1:]
	}
	return t
}

var (
	resourceRegex = regexp.MustCompile(`^\s*with\s+([^,]+),\s*$`)
	locationRegex = regexp.MustCompile(`^\s*on\s+(\S+)\s+line\s+(\d+)`)
)

func parseBlock(lines []string) (Diagnostic, bool) {
	var d Diagnostic
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "Error:"):
			d.Summary = strings.TrimSpace(strings.TrimPrefix(l, "Error:"))
		case d.Resource == "":
			if m := resourceRegex.FindStringSubmatch(l); m != nil {
				d.Resource = strings.TrimSpace(m[1])
			}
		}
		if m := locationRegex.FindStringSubmatch(l); m != nil {
			d.SourceFile = m[1]
			if n, err := strconv.Atoi(m[2]); err == nil {
				d.SourceLine = n
			}
		}
	}
	d.Raw = strings.Join(lines, "\n")
	if d.Summary == "" {
		return Diagnostic{}, false
	}
	return d, true
}

func parseLooseError(s string) (Diagnostic, bool) {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "Error:") {
			return Diagnostic{
				Summary: strings.TrimSpace(strings.TrimPrefix(t, "Error:")),
				Raw:     t,
			}, true
		}
	}
	return Diagnostic{}, false
}
