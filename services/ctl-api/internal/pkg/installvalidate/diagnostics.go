// Package installvalidate validates an install's desired configuration before
// it is persisted, surfacing structural problems as a set of diagnostics rather
// than a single opaque error. It is modelled loosely on HCL's diagnostics: each
// problem is a self-describing record (severity, code, summary, detail, the
// components involved) and a run aggregates every problem instead of failing on
// the first.
//
// Validation reasons about the install's DESIRED state (the component toggles in
// InstallConfig), never the actual deployment state. It is intended to gate the
// authoritative mutation paths — install-config sync and the component toggle
// endpoint — so a stored install can never reach a structurally invalid state.
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

// Diagnostic is a single self-describing validation problem.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	// Code is a stable, machine-readable identifier for the kind of problem.
	Code string `json:"code"`
	// Summary is a short, user-facing description of the problem.
	Summary string `json:"summary"`
	// Detail optionally explains how to resolve the problem.
	Detail string `json:"detail,omitempty"`
	// Components are the user-facing component names involved in the problem.
	Components []string `json:"components,omitempty"`
}

func (d Diagnostic) key() string {
	cs := append([]string(nil), d.Components...)
	sort.Strings(cs)
	return string(d.Severity) + "|" + d.Code + "|" + strings.Join(cs, ",")
}

// Diagnostics is an aggregated set of validation problems.
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

// Error renders the diagnostics as a single aggregated message so a non-empty
// Diagnostics value can be returned where an error is expected.
func (ds Diagnostics) Error() string {
	parts := make([]string, 0, len(ds))
	for _, d := range ds {
		parts = append(parts, d.Summary)
	}
	return strings.Join(parts, "; ")
}

// dedup collapses diagnostics that share a severity, code and component set,
// preserving first-seen order. Two rules can describe the same structural edge
// from opposite directions; dedup keeps a full sync from reporting it twice.
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
