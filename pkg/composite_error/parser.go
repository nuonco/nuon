package composite_error

import "context"

// ParseContext is a hierarchical, "/"-separated path identifying the producer
// of the raw error material being parsed. Examples:
//
//	"terraform"
//	"terraform/plan"
//	"terraform/apply"
//	"helm/install"
//	"helm/install/rbac"
//	"kubernetes/rollout"
//	"runner/job"
//	"build/helm"
//
// Parsers register against one or more ParseContexts and each entry matches
// itself plus all descendants. ParseContext is purely a dispatch concept and
// is never persisted on a CompositeError row — that's what Domain is for.
type ParseContext string

// ParseInput is what the dispatch pipeline hands to a parser. Not every field
// is required to be set — a parser examines what it cares about and returns
// Matched=false if the input doesn't apply.
type ParseInput struct {
	// Raw is the bytes the producer emitted (stderr, stdout, log buffer, …).
	Raw []byte

	// ExitCode of the producing process, if applicable.
	ExitCode int

	// GoErr carries the producing Go error. By convention the caller has
	// already stripped Temporal activity / workflow error envelopes via
	// `signal.HumanError` before handing it to the pipeline — parsers
	// should treat GoErr.Error() as the user-facing message and must not
	// re-walk Temporal wrappers themselves.
	GoErr error

	// Invocation describes who/what produced Raw.
	Invocation InvocationContext
}

// InvocationContext is metadata about who/what produced the raw input.
// Optional and best-effort — parsers must tolerate empty fields.
type InvocationContext struct {
	OrgID     string
	OwnerID   string
	OwnerType string

	ComponentID   string
	ComponentType string // "terraform_module", "helm_chart", "docker_build", ...

	InstallID string
	BuildID   string
	StepID    string

	// CloudPlatform is "aws" / "gcp" / "azure" if known.
	CloudPlatform string

	// Extra is a free-form bag for parser-specific hints.
	Extra map[string]any
}

// ParseResult is what a parser returns.
//
// Matched=false means "this parser does not claim the input"; Error/Source/Refs
// are ignored. Causes is a recursive list of sub-results that, when this result
// becomes the primary, are recorded as children in the cause graph.
type ParseResult struct {
	Matched bool
	Error   CompositeError
	Source  Source
	Refs    []Reference
	Causes  []ParseResult
}

// Parser turns raw error material into a typed CompositeError.
//
// Implementations live alongside their error type and register themselves
// into the parser catalog via init().
//
// Parsers MUST be best-effort, fast, and side-effect-free. The pipeline
// recovers panics, but a parser that panics on every call drowns the failure
// path in noise.
type Parser interface {
	// Name is the stable, human-readable identifier of the parser. Persisted
	// on the resulting CompositeError row's Source.ParserName for traceability.
	Name() string

	// Version is incremented when the parser's matching/extraction logic
	// changes in a way that would produce different output for the same input.
	Version() string

	// Contexts declares the ParseContext subtrees this parser opts into.
	// Each entry matches itself + descendants. An empty slice is a registration
	// error — every parser must opt in to at least one subtree.
	Contexts() []ParseContext

	// Parse examines the input and returns a ParseResult. If the input is not
	// a match, return ParseResult{Matched: false}.
	Parse(ctx context.Context, in ParseInput) ParseResult
}
