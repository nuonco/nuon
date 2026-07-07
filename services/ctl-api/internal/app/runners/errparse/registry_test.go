package errparse

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// fakeError is a trivial CompositeError used to assert which parser won.
type fakeError struct{ id string }

func (e fakeError) Error() string                    { return e.id }
func (e fakeError) Type() compositeerrors.Type       { return compositeerrors.Type(e.id) }
func (fakeError) Severity() compositeerrors.Severity { return compositeerrors.SeverityError }
func (fakeError) Sections() []compositeerrors.Section {
	return nil
}

// fakeParser is a configurable parser used to probe registry gating/ordering.
type fakeParser struct {
	id         string
	layer      Layer
	tools      []Tool
	signals    []string
	applicable bool
	ran        *bool
}

func (p fakeParser) Layer() Layer                  { return p.layer }
func (p fakeParser) Tools() []Tool                 { return p.tools }
func (p fakeParser) Signals() []string             { return p.signals }
func (p fakeParser) Applicable(*ParseContext) bool { return p.applicable }
func (p fakeParser) Parse(*ParseContext) compositeerrors.CompositeError {
	if p.ran != nil {
		*p.ran = true
	}
	return fakeError{id: p.id}
}

func TestRegistry_SignalGating(t *testing.T) {
	var ran bool
	r := &Registry{}
	r.Register(fakeParser{id: "p", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"AccessDenied"}, applicable: true, ran: &ran})

	// No signal present -> parser never runs.
	if ce := r.Parse(&ParseContext{Raw: "nothing to see", Tool: ToolTerraform}); ce != nil {
		t.Fatalf("expected nil, got %v", ce)
	}
	if ran {
		t.Fatal("parser ran despite no matching signal")
	}

	// Signal present -> parser runs.
	if ce := r.Parse(&ParseContext{Raw: "boom AccessDenied boom", Tool: ToolTerraform}); ce == nil {
		t.Fatal("expected a match once the signal is present")
	}
	if !ran {
		t.Fatal("expected parser to run when signal present")
	}
}

func TestRegistry_ToolBucketing(t *testing.T) {
	var tfRan, helmRan bool
	r := &Registry{}
	r.Register(fakeParser{id: "tf", layer: LayerTool, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true, ran: &tfRan})
	r.Register(fakeParser{id: "helm", layer: LayerTool, tools: []Tool{ToolHelm}, signals: []string{"boom"}, applicable: true, ran: &helmRan})

	ce := r.Parse(&ParseContext{Raw: "boom", Tool: ToolHelm})
	if ce == nil || ce.Error() != "helm" {
		t.Fatalf("expected helm parser to win, got %v", ce)
	}
	if tfRan {
		t.Fatal("terraform parser considered for a helm job")
	}
}

func TestRegistry_UnknownToolFailsOpen(t *testing.T) {
	r := &Registry{}
	r.Register(fakeParser{id: "tf", layer: LayerTool, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true})

	ce := r.Parse(&ParseContext{Raw: "boom", Tool: ToolUnknown})
	if ce == nil || ce.Error() != "tf" {
		t.Fatalf("expected tool-specific parser to run for unknown tool, got %v", ce)
	}
}

func TestRegistry_LayerOrdering(t *testing.T) {
	r := &Registry{}
	// Register generic first to prove ordering isn't registration order.
	r.Register(fakeParser{id: "generic", layer: LayerGeneric, signals: nil, applicable: true})
	r.Register(fakeParser{id: "provider", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true})

	ce := r.Parse(&ParseContext{Raw: "boom", Tool: ToolTerraform})
	if ce == nil || ce.Error() != "provider" {
		t.Fatalf("expected provider layer to win over generic, got %v", ce)
	}
}

func TestRegistry_AlwaysParserFallback(t *testing.T) {
	r := &Registry{}
	r.Register(fakeParser{id: "generic", layer: LayerGeneric, signals: nil, applicable: true})

	ce := r.Parse(&ParseContext{Raw: "any text at all", Tool: ToolTerraform})
	if ce == nil || ce.Error() != "generic" {
		t.Fatalf("expected always-parser fallback to match, got %v", ce)
	}
}

func TestRegistry_NotApplicableIsSkipped(t *testing.T) {
	r := &Registry{}
	r.Register(fakeParser{id: "p", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: false})

	if ce := r.Parse(&ParseContext{Raw: "boom", Tool: ToolTerraform}); ce != nil {
		t.Fatalf("expected nil when the only candidate is not applicable, got %v", ce)
	}
}

func TestRegistry_SameLayerTieBreakByRegistrationOrder(t *testing.T) {
	r := &Registry{}
	// Both provider-layer, both match, both applicable: first registered wins.
	r.Register(fakeParser{id: "first", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true})
	r.Register(fakeParser{id: "second", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true})

	// Run repeatedly; unknown tool exercises the fail-open path where bucket
	// collection order would otherwise be nondeterministic.
	for i := 0; i < 20; i++ {
		ce := r.Parse(&ParseContext{Raw: "boom", Tool: ToolUnknown})
		if ce == nil || ce.Error() != "first" {
			t.Fatalf("iteration %d: expected first-registered parser to win, got %v", i, ce)
		}
	}
}

func TestRegister_PanicsAfterBuild(t *testing.T) {
	r := &Registry{}
	r.Register(fakeParser{id: "p", layer: LayerProvider, tools: []Tool{ToolTerraform}, signals: []string{"boom"}, applicable: true})
	_ = r.Parse(&ParseContext{Raw: "boom", Tool: ToolTerraform}) // triggers build

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when registering after build")
		}
	}()
	r.Register(fakeParser{id: "late", layer: LayerProvider, signals: []string{"x"}})
}

func TestRegister_PanicsOnEmptySignal(t *testing.T) {
	r := &Registry{}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty signal string")
		}
	}()
	r.Register(fakeParser{id: "p", layer: LayerProvider, signals: []string{""}})
}

func TestAhoCorasick_MatchedSet(t *testing.T) {
	// Overlapping/suffix patterns exercise the failure links.
	m := newAhoCorasick([]string{"he", "she", "his", "hers"})
	got := m.matchedSet("ushers")
	want := []bool{true, true, false, true} // "he" and "hers" via "ushers"; "she" via "she"
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pattern %d: got %v want %v (full=%v)", i, got[i], want[i], got)
		}
	}

	none := newAhoCorasick([]string{"xyz"}).matchedSet("abcdef")
	if none[0] {
		t.Error("expected no match for absent pattern")
	}
}
