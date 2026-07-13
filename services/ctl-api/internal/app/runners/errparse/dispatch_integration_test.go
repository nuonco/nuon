package errparse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	awsparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/aws"
	genericparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/generic"
	helmparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/helm"
	tfparse "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse/terraform"
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

// TestDefaultRegistry_ToolLayerOrdering asserts the three-layer contract on a
// terraform job: a terraform diagnostic that no provider parser recognises
// yields the tool-layer parser (not the raw generic dump), while an AWS
// permission blob on the same job is still won by the provider layer.
func TestDefaultRegistry_ToolLayerOrdering(t *testing.T) {
	tfDiag, err := os.ReadFile(filepath.Join("terraform", "testdata", "invalid_reference.txt"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ce := errparse.Parse(&errparse.ParseContext{Raw: string(tfDiag), Tool: errparse.ToolTerraform})
	if _, ok := ce.(*tfparse.TerraformError); !ok {
		t.Fatalf("expected terraform tool-layer parser to win over generic, got %T (type %q)", ce, ce.Type())
	}

	awsBlob := "Error: creating S3 Bucket (acme): AccessDenied: User: " +
		"arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme"
	ce = errparse.Parse(&errparse.ParseContext{Raw: awsBlob, Tool: errparse.ToolTerraform})
	if _, ok := ce.(*awsparse.AWSPermissionError); !ok {
		t.Fatalf("expected AWS provider layer to win over terraform tool layer, got %T (type %q)", ce, ce.Type())
	}
}

// TestDefaultRegistry_StateLockBeatsTerraformCatchAll asserts that the
// state-lock parser wins over the generic terraform catch-all: both are
// candidates for a lock failure (each signals on the same diagnostic), and the
// state-lock parser's more-specific layer must break the tie.
func TestDefaultRegistry_StateLockBeatsTerraformCatchAll(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("terraform", "testdata", "state_lock.txt"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ce := errparse.Parse(&errparse.ParseContext{Raw: string(raw), Tool: errparse.ToolTerraform})
	if _, ok := ce.(*tfparse.StateLockError); !ok {
		t.Fatalf("expected state-lock parser to win over the terraform catch-all, got %T (type %q)", ce, ce.Type())
	}
}

// TestDefaultRegistry_HelmToolLayer asserts the same three-layer contract on a
// helm job: a recognised helm failure yields the tool-layer parser, while an
// AWS permission blob on the same job is still won by the provider layer.
func TestDefaultRegistry_HelmToolLayer(t *testing.T) {
	helmErr := "job step errored unable to execute job: unable to upgrade helm release: " +
		"cannot reuse a name that is still in use"
	ce := errparse.Parse(&errparse.ParseContext{Raw: helmErr, Tool: errparse.ToolHelm})
	if _, ok := ce.(*helmparse.HelmError); !ok {
		t.Fatalf("expected helm tool-layer parser to win over generic, got %T (type %q)", ce, ce.Type())
	}

	awsBlob := "unable to upgrade helm release: creating S3 Bucket (acme): AccessDenied: " +
		"User: arn:aws:iam::123:role/nuon-runner is not authorized to perform: " +
		"s3:CreateBucket on resource: arn:aws:s3:::acme"
	ce = errparse.Parse(&errparse.ParseContext{Raw: awsBlob, Tool: errparse.ToolHelm})
	if _, ok := ce.(*awsparse.AWSPermissionError); !ok {
		t.Fatalf("expected AWS provider layer to win over helm tool layer, got %T (type %q)", ce, ce.Type())
	}
}
