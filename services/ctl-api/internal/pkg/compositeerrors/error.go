// Package compositeerrors defines the typed, embedded error abstraction used
// across the Nuon platform.
//
// A CompositeError is a regular Go error (it satisfies the standard error
// interface) plus typed metadata — a discriminator, a severity, and an
// optional list of structured Sections — that lets the dashboard present a
// rich, opinionated view without losing the ability to be returned through
// the call stack like any other error.
//
// Composite errors are persisted by attaching a CompositeErrorData JSONB
// column to owner rows (component builds, sandbox runs, deploys, action runs, ..)
//
// New typed errors are added by writing a struct that implements
// CompositeError in its own subpackage (e.g. compositeerrors/terraform/,
// compositeerrors/validation/).
package compositeerrors

// Type is the discriminator string for a CompositeError implementation
// (e.g. "terraform.error", "validation").
type Type string

// Severity controls how the dashboard presents an error.
type Severity string

const (
	SeverityFatal   Severity = "fatal"
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// CompositeError is a typed, structured error. Implementations satisfy the
// standard error interface (Error() returns the headline message), plus
// metadata used to persist and present the error in the dashboard.
type CompositeError interface {
	error // headline — the one-line message users see

	Type() Type
	Severity() Severity

	// Sections returns optional structured detail (what / why / fix).
	// Returning nil is fine — the headline alone is a valid view.
	Sections() []Section
}

// Section is a heading + body attached to a CompositeError.
type Section struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}
