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
	HintSkipAutoRetry = "skip_auto_retry"

	HintRequeueAfter = "requeue_after"

	HintTerminal = "terminal"

	HintDocsURL = "docs_url"
)

func (h Hints) SkipAutoRetry() bool {
	v, _ := strconv.ParseBool(h[HintSkipAutoRetry])
	return v
}

func (h Hints) Terminal() bool {
	v, _ := strconv.ParseBool(h[HintTerminal])
	return v
}

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

func (h Hints) DocsURL() string {
	return h[HintDocsURL]
}

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

// why: NewHints returns an empty bag ready for the With* setters. Prefer the typed
// setters over raw map literals so canonical keys and value formats stay
// correct (a misspelled key or malformed value silently becomes a no-op).
func NewHints() Hints { return Hints{} }

func (h Hints) WithSkipAutoRetry() Hints {
	h[HintSkipAutoRetry] = "true"
	return h
}

func (h Hints) WithTerminal() Hints {
	h[HintTerminal] = "true"
	return h
}

func (h Hints) WithRequeueAfter(d time.Duration) Hints {
	if d < 0 {
		return h
	}
	h[HintRequeueAfter] = strconv.Itoa(int(d.Seconds()))
	return h
}

func (h Hints) WithDocsURL(url string) Hints {
	h[HintDocsURL] = url
	return h
}

type HintsProvider interface {
	Hints() Hints
}
