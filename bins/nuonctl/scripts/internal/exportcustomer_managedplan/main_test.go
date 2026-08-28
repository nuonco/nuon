package main

import (
	"testing"
)

func exporterSteps() []step {
	return []step{
		{ID: "create-1", JobType: "sandbox-terraform", JobOperation: "create-apply-plan", JobGroup: "sandbox", DependsOn: []string{"stale"}, PlanFromStep: "stale"},
		{ID: "apply-1", JobType: "sandbox-terraform", JobOperation: "apply-plan", JobGroup: "sandbox", DependsOn: []string{"stale"}, PlanFromStep: "stale"},
		{ID: "sync-1", JobType: "oci-sync", JobOperation: "exec", JobGroup: "sync"},
		{ID: "apply-other", JobType: "helm-chart-deploy", JobOperation: "apply-plan", JobGroup: "deploy"},
	}
}

func TestSelectStepsExplicitOrderAndChaining(t *testing.T) {
	selected, err := selectSteps(exporterSteps(), []string{"create-1", " apply-1 "})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "create-1" || selected[1].ID != "apply-1" {
		t.Fatalf("unexpected selection: %#v", selected)
	}
	if selected[0].DependsOn != nil || selected[0].PlanFromStep != "" {
		t.Fatalf("first step should have rebuilt empty dependencies: %#v", selected[0])
	}
	if len(selected[1].DependsOn) != 1 || selected[1].DependsOn[0] != "create-1" {
		t.Fatalf("apply should depend on prior selected step: %#v", selected[1])
	}
	if selected[1].PlanFromStep != "create-1" {
		t.Fatalf("compatible create/apply pair should chain plans: %#v", selected[1])
	}
}

func TestSelectStepsIncompatiblePairDoesNotChain(t *testing.T) {
	selected, err := selectSteps(exporterSteps(), []string{"sync-1", "apply-other"})
	if err != nil {
		t.Fatal(err)
	}
	if selected[1].PlanFromStep != "" {
		t.Fatalf("apply after mismatched job type must not chain: %#v", selected[1])
	}
	if len(selected[1].DependsOn) != 1 || selected[1].DependsOn[0] != "sync-1" {
		t.Fatalf("ordering dependency should still be rebuilt: %#v", selected[1])
	}
}

func TestSelectStepsUnknownAndEmpty(t *testing.T) {
	if _, err := selectSteps(exporterSteps(), []string{"missing"}); err == nil {
		t.Fatal("unknown job ID should error")
	}
	if _, err := selectSteps(exporterSteps(), []string{"", "  "}); err == nil {
		t.Fatal("empty selection should error")
	}
}
