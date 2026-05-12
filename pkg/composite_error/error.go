// Package composite_error defines the typed, embedded error abstraction used
// across the Nuon platform.
//
// A CompositeError is a regular Go error (it satisfies the standard error
// interface) plus typed metadata — a discriminator, a severity, and an
// optional list of structured Sections — that lets the dashboard present a
// rich, opinionated view without losing the ability to be returned through
// the call stack like any other error.
//
// Composite errors are persisted by attaching a CompositeErrorData JSONB
// column to owner rows (component builds, sandbox runs, deploys, action runs,
// …). A producer constructs a typed CompositeError inline at the failure
// site, hands it to New(), and assigns the resulting *CompositeErrorData to
// the owner. The headline message and Sections are frozen at write time —
// the read path is a plain JSONB unmarshal with no registry lookup.
//
// New typed errors are added by writing a struct that implements
// CompositeError in its own package (e.g. pkg/composite_error/terraform/,
// pkg/composite_error/validation/). There is no central registration step.
package composite_error

// Type is the discriminator string for a CompositeError implementation
// (e.g. "terraform.error", "validation"). Stored on the persisted
// CompositeErrorData record.
type Type string

// Severity controls how the dashboard presents an error. Producers set it on
// the CompositeError; the value is copied onto the CompositeErrorData record
// at construction time so reads never need to invoke the typed implementation.
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
//
// A CompositeError MUST be JSON round-trippable: marshalling and unmarshalling
// the implementing struct must yield an equivalent value. The
// CompositeErrorData's Data column holds exactly the JSON representation of
// the implementing struct.
type CompositeError interface {
	error // headline — the one-line message users see

	Type() Type
	Severity() Severity

	// Sections returns optional structured detail (what / why / fix).
	// Returning nil is fine — the headline alone is a valid view.
	Sections() []Section
}

// Section is a heading + markdown body attached to a CompositeError.
type Section struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}
