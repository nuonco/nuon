package compositeerrors

type Type string

// Severity controls how the dashboard presents an error.
type Severity string

const (
	SeverityFatal   Severity = "fatal"
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type CompositeError interface {
	error

	Type() Type
	Severity() Severity

	Sections() []Section
}

// SectionKind tells the renderer how to interpret a Section body, and, for
// security, whether the body is trusted. Untrusted content (raw tool output,
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

// why: MarkdownSection builds a trusted, markdown-rendered section. Only pass
// hand-authored, code-controlled content, never raw tool output.
func MarkdownSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionMarkdown}
}

func TextSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionText}
}

func CodeSection(heading, body string) Section {
	return Section{Heading: heading, Body: body, Kind: SectionCode}
}
