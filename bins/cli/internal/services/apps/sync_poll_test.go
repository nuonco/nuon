package apps

import (
	"strings"
	"testing"
)

func TestBuildOutcomesErr(t *testing.T) {
	allBuilt := []BuildOutcome{
		{ComponentID: "cmp1", ComponentName: "api", Status: buildOutcomeBuilt},
		{ComponentID: "cmp2", ComponentName: "worker", Status: buildOutcomeBuilt},
	}
	if err := buildOutcomesErr(allBuilt); err != nil {
		t.Fatalf("expected nil error when all builds complete, got %v", err)
	}

	mixed := []BuildOutcome{
		{ComponentID: "cmp1", ComponentName: "api", Status: buildOutcomeBuilt},
		{ComponentID: "cmp2", ComponentName: "worker", Status: buildOutcomeError},
		{ComponentID: "cmp3", ComponentName: "frontend", Status: buildOutcomePolicyFailed},
		{ComponentID: "cmp4", ComponentName: "job", Status: buildOutcomeTimeout},
	}
	err := buildOutcomesErr(mixed)
	if err == nil {
		t.Fatal("expected error when builds fail")
	}
	msg := err.Error()
	for _, want := range []string{
		"3 of 4 component build(s) did not complete",
		"worker (error)",
		"frontend (policy_failed)",
		"job (timeout)",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}
