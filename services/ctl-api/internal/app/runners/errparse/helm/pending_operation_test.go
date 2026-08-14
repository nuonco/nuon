package helm

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func parsePendingRaw(raw string) compositeerrors.CompositeError {
	return parsePendingOperation(&errparse.ParseContext{Raw: raw})
}

func TestParse_PendingOperation(t *testing.T) {
	ce := parsePendingRaw(readFixture(t, "pending_operation.txt"))
	if ce == nil {
		t.Fatal("expected a composite error, got nil")
	}
	e, ok := ce.(*PendingOperationError)
	if !ok {
		t.Fatalf("expected *PendingOperationError, got %T (type %q)", ce, ce.Type())
	}
	if ce.Type() != HelmPendingOperationType {
		t.Errorf("type = %q, want %q", ce.Type(), HelmPendingOperationType)
	}
	if ce.Severity() != compositeerrors.SeverityError {
		t.Errorf("severity = %q", ce.Severity())
	}
	if e.Status != "pending-upgrade" {
		t.Errorf("status = %q, want pending-upgrade", e.Status)
	}
	if !strings.Contains(ce.Error(), "pending-upgrade") {
		t.Errorf("headline should name the status: %q", ce.Error())
	}

	hp, ok := ce.(compositeerrors.HintsProvider)
	if !ok {
		t.Fatal("pending operation error should provide hints")
	}
	if !hp.Hints().SkipAutoRetry() {
		t.Error("a pending release cannot be cleared by a blind retry; expected skip_auto_retry")
	}

	headings := map[string]string{}
	for _, s := range ce.Sections() {
		headings[s.Heading] = s.Body
	}
	fix, ok := headings["How to fix"]
	if !ok {
		t.Fatal("expected a How to fix section")
	}
	if !strings.Contains(fix, "recover-helm-release") {
		t.Errorf("remediation should mention the CLI command: %q", fix)
	}
	if !strings.Contains(fix, "Recover Helm release") {
		t.Errorf("remediation should mention the dashboard action: %q", fix)
	}
	if _, ok := headings["Output"]; !ok {
		t.Error("expected an Output section")
	}
}

// The helm SDK wording carries no status, so the parser still has to fire and
// degrade the headline gracefully rather than claim a status it did not see.
func TestParse_PendingOperationFromSDKWording(t *testing.T) {
	ce := parsePendingRaw(readFixture(t, "pending_operation_sdk.txt"))
	e, ok := ce.(*PendingOperationError)
	if !ok {
		t.Fatalf("expected *PendingOperationError, got %T", ce)
	}
	if e.Status != "" {
		t.Errorf("status = %q, want empty when the output does not name one", e.Status)
	}
	if strings.Contains(ce.Error(), "pending-") {
		t.Errorf("headline must not invent a status: %q", ce.Error())
	}
}

// "cannot reuse a name that is still in use" is ambiguous: helm emits it for a
// pending release AND for a collision with a release Nuon does not own. Recovery
// is wrong advice for the latter, so it must stay on helm.name_in_use.
func TestParse_ReuseNameIsNotClassifiedAsPending(t *testing.T) {
	raw := readFixture(t, "reuse_name.txt")
	if ce := parsePendingRaw(raw); ce != nil {
		t.Fatalf("pending parser must defer on an ambiguous name collision, got %T", ce)
	}
	e := helmErr(t, parse(raw))
	if e.Type() != HelmNameInUseType {
		t.Errorf("type = %q, want %q", e.Type(), HelmNameInUseType)
	}
}

// Guard that the split does not swallow an unrelated helm failure.
func TestParse_OrdinaryHelmErrorIsNotPending(t *testing.T) {
	if ce := parsePendingRaw(readFixture(t, "immutable_field.txt")); ce != nil {
		t.Fatalf("pending parser must defer on an ordinary helm failure, got %T", ce)
	}
}
