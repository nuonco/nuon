package installvalidate

import (
	"sort"
	"strings"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Diagnostic struct {
	Severity   Severity `json:"severity"`
	Code       string   `json:"code"`
	Summary    string   `json:"summary"`
	Detail     string   `json:"detail,omitempty"`
	Components []string `json:"components,omitempty"`
}

func (d Diagnostic) key() string {
	cs := append([]string(nil), d.Components...)
	sort.Strings(cs)
	return string(d.Severity) + "|" + d.Code + "|" + strings.Join(cs, ",")
}

type Diagnostics []Diagnostic

func (ds Diagnostics) HasErrors() bool {
	for _, d := range ds {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

func (ds Diagnostics) Errors() Diagnostics {
	var out Diagnostics
	for _, d := range ds {
		if d.Severity == SeverityError {
			out = append(out, d)
		}
	}
	return out
}

func (ds Diagnostics) Error() string {
	parts := make([]string, 0, len(ds))
	for _, d := range ds {
		parts = append(parts, d.Summary)
	}
	return strings.Join(parts, "; ")
}

func (ds Diagnostics) dedup() Diagnostics {
	seen := make(map[string]struct{}, len(ds))
	out := make(Diagnostics, 0, len(ds))
	for _, d := range ds {
		k := d.key()
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, d)
	}
	return out
}
