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

// SectionKind tells the renderer how to interpret a Section body, and — for
// security — whether the body is trusted. Untrusted content (raw tool output,
// values extracted from an error) must never be rendered as markdown: the
// dashboard's markdown pipeline enables raw HTML and runs custom-component
// extraction over the string, so a crafted payload could escape a code fence
// and inject content. Use SectionText/SectionCode for anything derived from
// tool output, and reserve SectionMarkdown for hand-authored, trusted prose.
type SectionKind string

const (
	// SectionMarkdown renders the body as markdown. Only for trusted, code-
	// authored content. It is also the assumed kind for legacy records that
	// predate this field.
	SectionMarkdown SectionKind = "markdown"
	// SectionText renders the body as escaped plain text, preserving
	// whitespace. Safe for untrusted single- or multi-line values.
	SectionText SectionKind = "text"
	// SectionCode renders the body as an escaped monospace code block. Safe for
	// untrusted raw output (terraform/helm logs, an AWS response, ...).
	SectionCode SectionKind = "code"
)

// Section is a heading + body attached to a CompositeError. Kind controls how
// the body is rendered and whether it is treated as trusted (see SectionKind).
type Section struct {
	Heading string      `json:"heading"`
	Body    string      `json:"body"`
	Kind    SectionKind `json:"kind,omitempty"`
}

// MarkdownSection builds a trusted, markdown-rendered section. Only pass
// hand-authored, code-controlled content — never raw tool output.
func MarkdownSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionMarkdown}
}

// TextSection builds an escaped plain-text section. Safe for untrusted values.
func TextSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionText}
}

// CodeSection builds an escaped monospace code section. Safe for untrusted raw
// output; do not wrap the body in markdown fences.
func CodeSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionCode}
}
