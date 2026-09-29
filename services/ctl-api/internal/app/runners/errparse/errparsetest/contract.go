// why: Package errparsetest provides a shared, table-driven contract runner for
// errparse parsers. It dispatches fixtures through the real default registry
// (the same path production uses) and enforces the invariants every parser must
// satisfy: the expected type wins for a given input, and the resulting record
// survives the GORM persistence round-trip as valid JSON.
//
// Overlap and layer-ordering cases are only meaningful when every competing
// parser is registered, so full-registry contract runs must live in an external
// test package (for example package aws_test) that blank-imports errparse/all.
// An internal parser test (package aws) cannot blank-import errparse/all,
// because all imports every parser package and would form a cycle; run from
// there it would only see its own parser's registration, making overlap cases
// pass spuriously. errparsetest itself does not import errparse/all, both to
// avoid that cycle and to stay usable for parser-local checks.
package errparsetest

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type Case struct {
	Name     string
	Raw      string
	Tool     errparse.Tool
	Provider errparse.Provider
	WantType compositeerrors.Type
}

func Run(t *testing.T, cases []Case) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			pc := &errparse.ParseContext{Raw: tc.Raw, Tool: tc.Tool}
			if tc.Provider != errparse.ProviderUnknown {
				provider := tc.Provider
				pc.ResolveProvider = func() errparse.Provider { return provider }
			}

			ce := errparse.Parse(pc)
			if tc.WantType == "" {
				if ce != nil {
					t.Fatalf("expected no match, got %T (type %q)", ce, ce.Type())
				}
				return
			}
			if ce == nil {
				t.Fatalf("expected type %q, got no match", tc.WantType)
			}
			if ce.Type() != tc.WantType {
				t.Fatalf("winning type = %q, want %q (parser %T)", ce.Type(), tc.WantType, ce)
			}

			assertPersistable(t, ce)
		})
	}
}

func assertPersistable(t *testing.T, ce compositeerrors.CompositeError) {
	t.Helper()

	rec, err := compositeerrors.New(ce)
	if err != nil {
		t.Fatalf("compositeerrors.New(%T): %v", ce, err)
	}

	wantData, err := json.Marshal(ce)
	if err != nil {
		t.Fatalf("marshal %T payload: %v", ce, err)
	}
	if !bytes.Equal(rec.Data, wantData) {
		t.Fatalf("record Data = %s, want %s", rec.Data, wantData)
	}

	val, err := rec.Value()
	if err != nil {
		t.Fatalf("Value(): %v", err)
	}
	raw, ok := val.([]byte)
	if !ok {
		t.Fatalf("Value() returned %T, want []byte", val)
	}
	if !json.Valid(raw) {
		t.Fatalf("persisted record is not valid JSON: %s", raw)
	}

	var round compositeerrors.CompositeErrorData
	if err := round.Scan(raw); err != nil {
		t.Fatalf("Scan(): %v", err)
	}
	if round.Type != ce.Type() {
		t.Fatalf("round-trip type = %q, want %q", round.Type, ce.Type())
	}
	if !bytes.Equal(round.Data, rec.Data) {
		t.Fatalf("round-trip Data = %s, want %s", round.Data, rec.Data)
	}
}
