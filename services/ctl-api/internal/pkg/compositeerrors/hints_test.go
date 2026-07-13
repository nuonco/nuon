package compositeerrors

import (
	"testing"
	"time"
)

func TestHintsAccessors(t *testing.T) {
	t.Run("skip auto retry", func(t *testing.T) {
		if !(Hints{HintSkipAutoRetry: "true"}).SkipAutoRetry() {
			t.Fatal("expected SkipAutoRetry true")
		}
		if (Hints{HintSkipAutoRetry: "false"}).SkipAutoRetry() {
			t.Fatal("expected SkipAutoRetry false")
		}
		if (Hints{}).SkipAutoRetry() {
			t.Fatal("expected SkipAutoRetry false when unset")
		}
		if (Hints{HintSkipAutoRetry: "garbage"}).SkipAutoRetry() {
			t.Fatal("expected SkipAutoRetry false on malformed value")
		}
	})

	t.Run("terminal", func(t *testing.T) {
		if !(Hints{HintTerminal: "true"}).Terminal() {
			t.Fatal("expected Terminal true")
		}
		if (Hints{}).Terminal() {
			t.Fatal("expected Terminal false when unset")
		}
	})

	t.Run("requeue after", func(t *testing.T) {
		d, ok := (Hints{HintRequeueAfter: "300"}).RequeueAfter()
		if !ok || d != 5*time.Minute {
			t.Fatalf("expected 5m, got %v ok=%v", d, ok)
		}
		if _, ok := (Hints{}).RequeueAfter(); ok {
			t.Fatal("expected ok=false when unset")
		}
		if _, ok := (Hints{HintRequeueAfter: "nope"}).RequeueAfter(); ok {
			t.Fatal("expected ok=false on malformed value")
		}
		if _, ok := (Hints{HintRequeueAfter: "-5"}).RequeueAfter(); ok {
			t.Fatal("expected ok=false on negative value")
		}
	})

	t.Run("docs url", func(t *testing.T) {
		if got := (Hints{HintDocsURL: "https://x"}).DocsURL(); got != "https://x" {
			t.Fatalf("unexpected docs url %q", got)
		}
		if got := (Hints{}).DocsURL(); got != "" {
			t.Fatalf("expected empty docs url, got %q", got)
		}
	})
}

// hintedError is a minimal CompositeError that advertises hints, used to verify
// New() captures them.
type hintedError struct{ hints Hints }

func (hintedError) Error() string      { return "boom" }
func (hintedError) Type() Type         { return "test.hinted" }
func (hintedError) Severity() Severity { return SeverityError }
func (hintedError) Sections() []Section {
	return nil
}
func (e hintedError) Hints() Hints { return e.hints }

// plainError is a CompositeError without hints.
type plainError struct{}

func (plainError) Error() string       { return "plain" }
func (plainError) Type() Type          { return "test.plain" }
func (plainError) Severity() Severity  { return SeverityWarning }
func (plainError) Sections() []Section { return nil }

func TestHintsWriters(t *testing.T) {
	h := NewHints().WithSkipAutoRetry().WithTerminal().WithDocsURL("https://docs.nuon.co/x").WithRequeueAfter(5 * time.Minute)

	if !h.SkipAutoRetry() {
		t.Error("expected skip_auto_retry set")
	}
	if !h.Terminal() {
		t.Error("expected terminal set")
	}
	if h.DocsURL() != "https://docs.nuon.co/x" {
		t.Errorf("docs_url = %q", h.DocsURL())
	}
	if d, ok := h.RequeueAfter(); !ok || d != 5*time.Minute {
		t.Errorf("requeue_after = %v ok=%v", d, ok)
	}

	if _, ok := NewHints().WithRequeueAfter(-1 * time.Second).RequeueAfter(); ok {
		t.Error("a negative duration must leave requeue_after unset")
	}
}

func TestNewCapturesVersionHintsAndSource(t *testing.T) {
	t.Run("captures version and hints", func(t *testing.T) {
		d, err := New(hintedError{hints: Hints{HintSkipAutoRetry: "true"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.Version != SchemaVersion {
			t.Fatalf("expected version %d, got %d", SchemaVersion, d.Version)
		}
		if !d.Hints.SkipAutoRetry() {
			t.Fatal("expected captured hint SkipAutoRetry true")
		}
	})

	t.Run("no hints when provider absent", func(t *testing.T) {
		d, err := New(plainError{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.Hints != nil {
			t.Fatalf("expected nil hints, got %v", d.Hints)
		}
	})

	t.Run("empty hints not stored", func(t *testing.T) {
		d, err := New(hintedError{hints: Hints{}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.Hints != nil {
			t.Fatalf("expected nil hints for empty bag, got %v", d.Hints)
		}
	})

	t.Run("WithSource records provenance", func(t *testing.T) {
		d, err := New(plainError{}, WithSource("runner_job_execution_results", "rje_123"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d.SourceType != "runner_job_execution_results" || d.SourceID != "rje_123" {
			t.Fatalf("unexpected source: %s/%s", d.SourceType, d.SourceID)
		}
	})
}
