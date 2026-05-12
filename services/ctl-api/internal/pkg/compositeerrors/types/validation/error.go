// Package validation provides a generic validation CompositeError used when
// an owner's preconditions or inputs are invalid (a bad field, an entity in a
// wrong state, etc.). Producers construct it inline at the failure site:
//
//	build.CompositeError = compositeerrors.New(&validation.Error{
//	    Field:   "component.status",
//	    Message: "component must be active before a build can run",
//	    Got:     string(comp.Status),
//	})
package validation

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// Error is a typed validation failure. Implements the standard error
// interface plus compositeerrors.CompositeError so it can be returned from any
// func() error and wrapped at the persistence boundary via compositeerrors.New.
type Error struct {
	// Field is the dotted path to the offending field or condition
	// (e.g. "component.status", "manifest.name").
	Field string `json:"field"`

	// Message is the user-facing explanation of what's wrong.
	Message string `json:"message"`

	// Got is the actual value observed, when meaningful. Optional.
	Got string `json:"got,omitempty"`
}

var _ compositeerrors.CompositeError = (*Error)(nil)

func (e *Error) Error() string                      { return fmt.Sprintf("%s: %s", e.Field, e.Message) }
func (e *Error) Type() compositeerrors.Type         { return "validation" }
func (e *Error) Severity() compositeerrors.Severity { return compositeerrors.SeverityError }
func (e *Error) Sections() []compositeerrors.Section {
	if e.Got == "" {
		return nil
	}
	return []compositeerrors.Section{{Heading: "Got", Body: e.Got}}
}
