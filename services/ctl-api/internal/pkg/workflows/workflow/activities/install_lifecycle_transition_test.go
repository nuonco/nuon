package activities

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func provisionWorkflow(status app.Status, stackOnly bool) *app.Workflow {
	metadata := pgtype.Hstore{}
	if stackOnly {
		metadata[app.WorkflowMetadataKeyStackOnly] = generics.ToPtr("true")
	}
	return &app.Workflow{
		Type:     app.WorkflowTypeProvision,
		Metadata: metadata,
		Status:   app.CompositeStatus{Status: status},
	}
}

// A stack-only provision deliberately stops before the sandbox and components,
// so reporting it as provisioned would mark an install complete while it has no
// sandbox and zero component deploys.
func TestInstallLifecycleTransitionStackOnlyProvision(t *testing.T) {
	a := &Activities{}

	got := a.installLifecycleTransition(provisionWorkflow(app.StatusSuccess, true))
	if got == nil {
		t.Fatal("expected a lifecycle phase for a provision workflow")
	}
	if got.Phase != lifecyclephase.Provisioning {
		t.Fatalf("stack-only provision: phase = %q, want %q", got.Phase, lifecyclephase.Provisioning)
	}
}

func TestInstallLifecycleTransitionFullProvision(t *testing.T) {
	a := &Activities{}

	got := a.installLifecycleTransition(provisionWorkflow(app.StatusSuccess, false))
	if got == nil {
		t.Fatal("expected a lifecycle phase for a provision workflow")
	}
	if got.Phase != lifecyclephase.Provisioned {
		t.Fatalf("full provision: phase = %q, want %q", got.Phase, lifecyclephase.Provisioned)
	}
}

// A failed stack-only provision is still a failure: it must not be reported as
// the healthy "waiting to provision the sandbox" hold.
func TestInstallLifecycleTransitionStackOnlyFailureKeepsFailureDescription(t *testing.T) {
	a := &Activities{}

	for _, status := range []app.Status{app.StatusError, app.StatusCancelled} {
		got := a.installLifecycleTransition(provisionWorkflow(status, true))
		if got == nil {
			t.Fatalf("%s: expected a lifecycle phase", status)
		}
		if got.Phase != lifecyclephase.Provisioned {
			t.Fatalf("%s: phase = %q, want %q", status, got.Phase, lifecyclephase.Provisioned)
		}
		if got.Description != "Provision workflow failed" {
			t.Fatalf("%s: description = %q, want the failure description", status, got.Description)
		}
	}
}
