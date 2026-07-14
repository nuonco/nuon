package errparse

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func nilParse(*ParseContext) compositeerrors.CompositeError { return nil }

func TestNewParser_Defaults(t *testing.T) {
	p := NewParser(LayerTool, nilParse, WithSignals("boom"))

	if p.Layer() != LayerTool {
		t.Errorf("layer = %d, want %d", p.Layer(), LayerTool)
	}
	if p.Tools() != nil {
		t.Errorf("tools = %v, want nil (tool-agnostic)", p.Tools())
	}
	if got := p.Signals(); len(got) != 1 || got[0] != "boom" {
		t.Errorf("signals = %v", got)
	}
	if !p.Applicable(&ParseContext{}) {
		t.Error("Applicable should default to true when no providers are set")
	}
}

func TestNewParser_AlwaysCandidate(t *testing.T) {
	p := NewParser(LayerGeneric, nilParse, AlwaysCandidate())
	if p.Signals() != nil {
		t.Errorf("signals = %v, want nil for an always-candidate parser", p.Signals())
	}
}

func TestNewParser_ProviderGate(t *testing.T) {
	p := NewParser(LayerProvider, nilParse, WithSignals("boom"), WithProviders(ProviderAWS))

	cases := []struct {
		provider Provider
		want     bool
	}{
		{ProviderAWS, true},
		{ProviderUnknown, true}, // fails open
		{ProviderAzure, false},
		{ProviderGCP, false},
	}
	for _, tc := range cases {
		ctx := &ParseContext{ResolveProvider: func() Provider { return tc.provider }}
		if got := p.Applicable(ctx); got != tc.want {
			t.Errorf("Applicable(provider=%q) = %v, want %v", tc.provider, got, tc.want)
		}
	}
}

func TestNewParser_ProviderGateNotResolvedWithoutCall(t *testing.T) {
	// A parser with no provider gate must never trigger provider resolution.
	resolved := false
	p := NewParser(LayerTool, nilParse, WithSignals("boom"))
	ctx := &ParseContext{ResolveProvider: func() Provider { resolved = true; return ProviderAWS }}
	if !p.Applicable(ctx) {
		t.Error("expected always-applicable without a provider gate")
	}
	if resolved {
		t.Error("provider resolver called for a parser with no provider gate")
	}
}

func TestNewParser_Panics(t *testing.T) {
	cases := map[string]func(){
		"nil parse":        func() { NewParser(LayerTool, nil, WithSignals("x")) },
		"no signal choice": func() { NewParser(LayerTool, nilParse) },
		"both signal choices": func() {
			NewParser(LayerTool, nilParse, WithSignals("x"), AlwaysCandidate())
		},
		"empty signals":    func() { WithSignals() },
		"empty providers":  func() { WithProviders() },
		"unknown provider": func() { WithProviders(ProviderUnknown) },
		"nil option":       func() { NewParser(LayerTool, nilParse, nil) },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %q", name)
				}
			}()
			fn()
		})
	}
}
