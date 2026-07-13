package compositeerrors

import (
	"strconv"
	"time"
)

// Hints is an open annotation bag carried on a CompositeError. Unlike Data
// (which is the typed, per-error-type description of WHAT the error is), Hints
// describes HOW the platform should handle or present the error. It is
// cross-cutting and not tied to any single error type's schema.
//
// Values are always strings to keep a single, unambiguous wire format inside
// the JSONB payload (a map[string]any would round-trip numbers as float64 and
// drift). Consumers read canonical keys through the typed accessors below,
// which own the coercion. Non-scalar values belong in Data, not Hints.
//
// The bag is open, but the keys a consumer ACTS ON are a documented, closed
// set (the Hint* constants). An error may attach arbitrary annotation keys,
// but only canonical keys carry behavior.
type Hints map[string]string

const (
	// HintSkipAutoRetry ("true"): the orchestrator should not auto-retry this
	// failure; park the step for manual retry instead. Used for errors that
	// won't resolve by retrying (e.g. a missing IAM permission).
	HintSkipAutoRetry = "skip_auto_retry"

	// HintRequeueAfter (integer seconds, e.g. "300"): the orchestrator should
	// back off for the given duration before retrying rather than parking.
	// Used for transient, time-bounded failures (e.g. quota/throttle).
	HintRequeueAfter = "requeue_after"

	// HintTerminal ("true"): the failure is not retryable at all (auto or
	// manual).
	HintTerminal = "terminal"

	// HintDocsURL: a documentation link the UI may surface as "learn more".
	HintDocsURL = "docs_url"
)

// SkipAutoRetry reports whether the orchestrator should skip auto-retries.
func (h Hints) SkipAutoRetry() bool {
	v, _ := strconv.ParseBool(h[HintSkipAutoRetry])
	return v
}

// Terminal reports whether the failure is not retryable at all.
func (h Hints) Terminal() bool {
	v, _ := strconv.ParseBool(h[HintTerminal])
	return v
}

// RequeueAfter returns the back-off duration and whether it was set. A missing
// or malformed value returns (0, false).
func (h Hints) RequeueAfter() (time.Duration, bool) {
	s, ok := h[HintRequeueAfter]
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, false
	}
	return time.Duration(n) * time.Second, true
}

// DocsURL returns the documentation link, or "" when unset.
func (h Hints) DocsURL() string {
	return h[HintDocsURL]
}

// Clone returns a shallow copy of the bag, or nil when empty. Use it before
// mutating a Hints value that may be shared (e.g. a package-level default).
func (h Hints) Clone() Hints {
	if len(h) == 0 {
		return nil
	}
	out := make(Hints, len(h))
	for k, v := range h {
		out[k] = v
	}
	return out
}

// NewHints returns an empty bag ready for the With* setters. Prefer the typed
// setters over raw map literals so canonical keys and value formats stay
// correct (a misspelled key or malformed value silently becomes a no-op).
func NewHints() Hints { return Hints{} }

// WithSkipAutoRetry marks the failure so the orchestrator parks the step for
// manual retry instead of auto-retrying.
func (h Hints) WithSkipAutoRetry() Hints {
	h[HintSkipAutoRetry] = "true"
	return h
}

// WithTerminal marks the failure as not retryable at all.
func (h Hints) WithTerminal() Hints {
	h[HintTerminal] = "true"
	return h
}

// WithRequeueAfter sets the back-off before retrying. A negative duration is
// ignored so the bag never carries a malformed value.
func (h Hints) WithRequeueAfter(d time.Duration) Hints {
	if d < 0 {
		return h
	}
	h[HintRequeueAfter] = strconv.Itoa(int(d.Seconds()))
	return h
}

// WithDocsURL attaches a documentation link the UI may surface as "learn more".
func (h Hints) WithDocsURL(url string) Hints {
	h[HintDocsURL] = url
	return h
}

// HintsProvider is an optional capability implemented by typed CompositeError
// values that want to attach hints. New() captures Hints() at write time, the
// same way it captures Sections().
type HintsProvider interface {
	Hints() Hints
}
