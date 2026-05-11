// Package unknown is the always-last fallback CompositeError type, shipped as
// a builtin so that every recorded incident has a typed, catalog-registered
// representation even when no parser matches.
package unknown

import (
	"context"
	"fmt"
	"strings"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

// Type is the catalog identifier for the unknown error.
const Type composite_error.Type = "unknown_error"

// Error is the typed representation of an unclassified error. All fields are
// best-effort: even an empty value renders to a usable (if generic) view.
type Error struct {
	// Message is the cleaned outermost error string. Used as the headline.
	Message string `json:"message,omitempty"`

	// Details is the full (capped) raw error text. Rendered as an "Error
	// details" section when it carries information beyond Message — i.e.
	// multi-line stderr, wrapped Go error chains, etc.
	Details string `json:"details,omitempty"`

	// ExitCode of the producing process, when known.
	ExitCode *int `json:"exit_code,omitempty"`
}

var _ composite_error.CompositeError = (*Error)(nil)

func (e *Error) Type() composite_error.Type         { return Type }
func (e *Error) Domain() composite_error.Domain     { return composite_error.DomainNuon }
func (e *Error) Severity() composite_error.Severity { return composite_error.SeverityError }

func (e *Error) Render(_ context.Context) composite_error.Render {
	title := e.Message
	if title == "" {
		title = "An unknown error occurred"
	}

	r := composite_error.Render{
		Title: title,
	}
	if details := strings.TrimSpace(e.Details); details != "" && details != strings.TrimSpace(e.Message) {
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Error details",
			Body:    details,
		})
	}
	if e.ExitCode != nil {
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Exit code",
			Body:    fmt.Sprintf("%d", *e.ExitCode),
		})
	}
	return r
}
