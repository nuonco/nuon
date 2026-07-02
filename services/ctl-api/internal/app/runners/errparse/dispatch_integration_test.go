package errparse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	awsparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/aws"
	genericparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
)

// This external test package imports both parser packages so their init()
// registrations land in the default registry together, matching the chokepoint
// wiring. It asserts the layer contract that only holds once specific and
// generic parsers coexist: a specific provider-layer parser wins, and the
// generic fallback catches everything else.

func TestDefaultRegistry_SpecificBeatsGeneric(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("aws", "testdata", "terraform_apply_access_denied.txt"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	ce := errparse.Parse(&errparse.ParseContext{Raw: string(raw)})
	if ce == nil {
		t.Fatal("expected a match")
	}
	if _, ok := ce.(*awsparse.AWSPermissionError); !ok {
		t.Fatalf("expected AWS permission error to win over generic, got %T (type %q)", ce, ce.Type())
	}
}

func TestDefaultRegistry_GenericCatchesUnclassified(t *testing.T) {
	ce := errparse.Parse(&errparse.ParseContext{Raw: "helm upgrade failed: context deadline exceeded"})
	if ce == nil {
		t.Fatal("expected the generic fallback to match")
	}
	if _, ok := ce.(*genericparse.GenericError); !ok {
		t.Fatalf("expected generic fallback, got %T (type %q)", ce, ce.Type())
	}
}
