// Package unknown_error is the always-last fallback CompositeError type.
//
// When no parser matches the raw error material, the pipeline produces an
// instance of this type so every recorded incident has a typed,
// catalog-registered representation.
package unknown_error

import (
	"context"
	"fmt"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

// Type is the catalog identifier for this error.
const Type composite_error.Type = "unknown_error"

// Error is the typed representation of an unclassified error.
//
// All fields are best-effort: they are populated from whatever the producer
// gave us. Even an empty value renders to a usable (if generic) view.
type Error struct {
	// Message is the cleaned outermost error string (HumanError() output, or
	// the raw stderr first line). Used as the headline.
	Message string `json:"message,omitempty"`

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
		Title:   title,
		Summary: "We weren't able to classify this error. The raw output is preserved in the error source.",
	}
	if e.ExitCode != nil {
		r.Sections = append(r.Sections, composite_error.RenderSection{
			Heading: "Exit code",
			Body:    fmt.Sprintf("%d", *e.ExitCode),
		})
	}
	return r
}
