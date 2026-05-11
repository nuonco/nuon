package composite_error

// DirectiveKind aligns with WorkflowStep.ResultDirective so the conductor can
// apply an error-driven override without translation.
type DirectiveKind string

const (
	// DirectiveNone is the zero value — "no opinion, defer to signal defaults".
	DirectiveNone DirectiveKind = ""

	DirectiveContinue      DirectiveKind = "continue"
	DirectiveStop          DirectiveKind = "stop"
	DirectiveRetry         DirectiveKind = "retry"
	DirectiveRetryGroup    DirectiveKind = "retry-group"
	DirectiveSkipGroup     DirectiveKind = "skip-group"
	DirectiveAwaitApproval DirectiveKind = "await-approval"
)

// Directive is the override an error type can apply to a step's outcome. The
// shape is intentional: future fields (MaxRetries, Backoff, …) can be added
// without breaking the capability interface.
type Directive struct {
	Kind DirectiveKind `json:"kind"`
}

// ErrorWithDirective is implemented by error types that override the conductor's
// default directive for a failed step.
//
// Returning Directive{Kind: DirectiveNone} (the zero value) means the type has
// no opinion and the signal's default applies.
type ErrorWithDirective interface {
	OverrideDirective() Directive
}

// ErrorWithDocsLink is implemented by error types that point at a docs page.
type ErrorWithDocsLink interface {
	DocsURL() string
}

// ErrorWithReferences is implemented by error types that declare static
// references in addition to anything the parser attaches dynamically.
type ErrorWithReferences interface {
	References() []Reference
}

// ErrorWithJSONSchema is implemented by error types that provide a JSON Schema
// for their Data payload, validated by helpers.Record() before persisting.
type ErrorWithJSONSchema interface {
	JSONSchema() []byte
}
