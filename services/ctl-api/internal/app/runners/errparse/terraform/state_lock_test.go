package terraform

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func parseStateLock(raw string) compositeerrors.CompositeError {
	return stateLockParser{}.Parse(&errparse.ParseContext{Raw: raw})
}

func TestParse_StateLock(t *testing.T) {
	ce := parseStateLock(readFixture(t, "state_lock.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e, ok := ce.(*StateLockError)
	if !ok {
		t.Fatalf("expected *StateLockError, got %T (type %q)", ce, ce.Type())
	}
	if ce.Type() != TerraformStateLockType {
		t.Errorf("type = %q, want %q", ce.Type(), TerraformStateLockType)
	}
	if ce.Severity() != compositeerrors.SeverityError {
		t.Errorf("severity = %q", ce.Severity())
	}
	if !strings.Contains(ce.Error(), "state lock") {
		t.Errorf("headline = %q", ce.Error())
	}

	hp, ok := ce.(compositeerrors.HintsProvider)
	if !ok {
		t.Fatal("state lock error should provide hints")
	}
	if !hp.Hints().SkipAutoRetry() {
		t.Error("a stale state lock cannot be cleared by a blind retry; expected skip_auto_retry")
	}

	headings := map[string]string{}
	for _, s := range ce.Sections() {
		headings[s.Heading] = s.Body
	}
	fix, ok := headings["How to fix"]
	if !ok {
		t.Fatal("expected a How to fix section")
	}
	if !strings.Contains(fix, "force-unlock") {
		t.Errorf("remediation should mention force-unlock: %q", fix)
	}
	if !strings.Contains(fix, "Unlock Terraform state") {
		t.Errorf("remediation should mention the dashboard unlock action: %q", fix)
	}
	if _, ok := headings["Output"]; !ok {
		t.Error("expected an Output section")
	}
	_ = e
}

// TestParse_OrdinaryErrorIsNotStateLock guards that the state-lock split does
// not swallow a normal terraform diagnostic.
func TestParse_OrdinaryErrorIsNotStateLock(t *testing.T) {
	raw := readFixture(t, "invalid_reference.txt")
	if ce := parseStateLock(raw); ce != nil {
		t.Fatalf("state-lock parser must defer on an ordinary error, got %T", ce)
	}
	if _, ok := parse(raw).(*TerraformError); !ok {
		t.Fatal("the terraform catch-all should classify an ordinary diagnostic")
	}
}
