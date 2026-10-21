package service

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

func TestPinOnboardingBranch(t *testing.T) {
	params := &helpers.CreateInstallParams{Name: "demo"}
	branchID := "brnch1"
	pinOnboardingBranch(params, &app.Onboarding{AppBranchID: &branchID})
	if params.AppBranchID != branchID {
		t.Fatalf("AppBranchID = %q, want %q", params.AppBranchID, branchID)
	}

	unbranched := &helpers.CreateInstallParams{Name: "demo"}
	pinOnboardingBranch(unbranched, &app.Onboarding{})
	if unbranched.AppBranchID != "" {
		t.Fatalf("AppBranchID = %q, want empty", unbranched.AppBranchID)
	}
}
