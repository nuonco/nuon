// Package composite_error defines the typed, catalogable error abstraction
// used across the Nuon platform.
//
// See specs/composite-errors.md for design rationale.
//
// At a high level:
//
//   - A CompositeError is a typed Go value (one struct per error type) that
//     knows how to render itself for users.
//   - Each type registers a factory in the in-memory catalog via init(),
//     mirroring how queue signals are registered.
//   - Parsers (also registered in the catalog) consume raw error material
//     (Temporal errors, runner stderr, terraform output, …) and emit typed
//     CompositeErrors via a dispatch pipeline.
//   - The conductor / step finalizer honors optional capability interfaces
//     on a CompositeError (see directive.go) to override step directives
//     (fail-fast, retry, downgrade, …).
//
// Phase 1 lands the abstraction and the catalog only; the helpers, GORM models,
// and parser implementations live alongside ctl-api.
package composite_error

import "context"

// Type is the catalog identifier for a CompositeError implementation.
// Stored on the persisted row and used to look up the factory in the catalog.
type Type string

// Domain is the broad classification of an error — what kind of error it is,
// not where it came from. Domain is flat (no hierarchy) and lives on the row
// for filtering, metrics, and UI grouping.
type Domain string

// Standard domains. New domains are introduced as needed; kept short and stable.
const (
	DomainNuon       Domain = "nuon"
	DomainTerraform  Domain = "terraform"
	DomainHelm       Domain = "helm"
	DomainKubernetes Domain = "kubernetes"
	DomainAWS        Domain = "aws"
	DomainGCP        Domain = "gcp"
	DomainAzure      Domain = "azure"
	DomainRunner     Domain = "runner"
)

// Severity controls how the UI presents an error and whether it is treated as
// a hard failure. The conductor reacts to OverrideDirective(), not Severity —
// severity is a UX concept.
type Severity string

const (
	SeverityFatal   Severity = "fatal"
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// CompositeError is the typed, in-memory view of a stored composite error.
//
// Implementations live alongside ctl-api under
// services/ctl-api/internal/app/composite_errors/types/<name>/ and register
// themselves into the catalog via init().
//
// A CompositeError MUST be JSON round-trippable: marshalling the implementing
// struct and unmarshalling into a fresh value produced by the catalog factory
// must yield an equivalent value. The persisted row's Data column holds
// exactly the JSON representation of the implementing struct.
type CompositeError interface {
	Type() Type
	Domain() Domain
	Severity() Severity

	// Render produces the user-facing view from the typed fields on the
	// implementing struct. Always called at read time. The TitleCached /
	// SummaryCached columns on the persisted row are denormalized copies
	// populated from Render() at write time and used only for admin
	// search / listing.
	Render(ctx context.Context) Render
}

// Render is the user-facing payload computed from a CompositeError's typed
// fields. Sections and UserActions are ordered.
type Render struct {
	Title       string          `json:"title"`
	Summary     string          `json:"summary,omitempty"`
	Sections    []RenderSection `json:"sections,omitempty"`
	UserActions []UserAction    `json:"user_actions,omitempty"`
}

// RenderSection is a heading + markdown body. Sections typically follow a
// what / why / how-to-fix structure.
type RenderSection struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// UserAction is a structured CTA the UI can render as a button or chip.
type UserAction struct {
	Kind  UserActionKind `json:"kind"`
	Label string         `json:"label"`
	Value string         `json:"value,omitempty"` // url, snippet, command, etc.
}

type UserActionKind string

const (
	UserActionKindLink    UserActionKind = "link"
	UserActionKindCopy    UserActionKind = "copy"
	UserActionKindCommand UserActionKind = "command"
)
