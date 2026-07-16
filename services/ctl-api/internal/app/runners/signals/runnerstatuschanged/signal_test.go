package runnerstatuschanged

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestLifecycleContext(t *testing.T) {
	s := &Signal{RunnerID: "runner", OrgID: "org", FromStatus: app.RunnerStatusOffline, ToStatus: app.RunnerStatusActive, Reason: "ready", RunnerGroupType: app.RunnerGroupTypeInstall, OwnerID: "install", OwnerType: "installs"}
	ctx := s.LifecycleContext()
	if ctx.Operation != "active" || ctx.OwnerID != "runner" || ctx.OwnerType != "runners" || ctx.InstallID == nil || *ctx.InstallID != "install" {
		t.Fatalf("unexpected lifecycle context: %#v", ctx)
	}
	if ctx.Metadata["from_status"] != "offline" || ctx.Metadata["to_status"] != "active" {
		t.Fatalf("unexpected metadata: %#v", ctx.Metadata)
	}
}

func TestValidatePayloadOnly(t *testing.T) {
	valid := &Signal{RunnerID: "offline-runner", OrgID: "org", FromStatus: app.RunnerStatusActive, ToStatus: app.RunnerStatusOffline, Reason: "disconnected", RunnerGroupType: app.RunnerGroupTypeInstall, OwnerID: "install", OwnerType: "installs"}
	if err := valid.Validate(nil); err != nil {
		t.Fatalf("valid payload: %v", err)
	}
	valid.RunnerID = ""
	if err := valid.Validate(nil); err == nil {
		t.Fatal("missing payload field should fail validation")
	}
	valid.RunnerID = "offline-runner"
	valid.Reason = ""
	if err := valid.Validate(nil); err != nil {
		t.Fatalf("empty reason should remain deliverable: %v", err)
	}
}
