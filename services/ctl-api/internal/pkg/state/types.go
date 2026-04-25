package state

import (
	"time"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
)

// StateManagerRequest is the input to the StateManager workflow.
type StateManagerRequest struct {
	InstallID string
	State     *StateManagerState
}

// StateManagerState is carried between continue-as-new cycles.
type StateManagerState struct {
	// LastModifiedAt tracks the last-known updated_at per partial.
	LastModifiedAt map[PartialName]time.Time

	// CachedState holds the last-generated full state.
	CachedState *pkgstate.State

	// LastGeneratedAt tracks when the state was last persisted.
	LastGeneratedAt time.Time

	// GenerationCount tracks total generations for diagnostics.
	GenerationCount int64
}

// Update handler names.
const (
	ForceRegenerateUpdateName = "force-regenerate"
	RegenerateUpdateName      = "regenerate"
	HintUpdateName            = "hint"
	FetchStateUpdateName      = "fetch-state"
	StatusQueryName           = "status"
	StopUpdateName            = "stop"
	RestartUpdateName         = "restart"
)

// ForceRegenerate — full rebuild of all partials.
type ForceRegenerateRequest struct{}
type ForceRegenerateResponse struct {
	State       *pkgstate.State
	GeneratedAt time.Time
}

// Regenerate — check all partials for changes, update stale ones.
type RegenerateRequest struct{}
type RegenerateResponse struct {
	State           *pkgstate.State
	UpdatedPartials []PartialName
	GeneratedAt     time.Time
}

// Hint — targeted regeneration based on what changed.
type HintRequest struct {
	HintType HintType
	EntityID string
}
type HintResponse struct {
	State           *pkgstate.State
	UpdatedPartials []PartialName
	GeneratedAt     time.Time
}

// FetchState — return the current cached state without regenerating.
type FetchStateRequest struct{}
type FetchStateResponse struct {
	State       *pkgstate.State
	GeneratedAt time.Time
}

// Status — query workflow metadata (lightweight).
type StatusResponse struct {
	Ready           bool
	LastGeneratedAt time.Time
	GenerationCount int64
	LastModifiedAt  map[PartialName]time.Time
}

// Stop — terminate the workflow.
type StopRequest struct{}
type StopResponse struct{}

// Restart — trigger continue-as-new.
type RestartRequest struct{}
type RestartResponse struct{}
