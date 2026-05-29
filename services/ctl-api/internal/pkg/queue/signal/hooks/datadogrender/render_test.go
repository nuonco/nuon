package datadogrender

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
)

func TestBuild_WorkflowFailedSetsErrorAlertAndStatusTag(t *testing.T) {
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflow,
		Transition: slackrender.TransitionFailed,
		OrgID:      "org-1",
		Workflow: slackrender.WorkflowRef{
			ID:        "wf-1",
			Type:      slackrender.WorkflowTypeProvision,
			OwnerID:   "ins-1",
			OwnerType: slackrender.OwnerTypeInstalls,
			OwnerName: "acme-prod",
		},
		Outcome: &slackrender.Outcome{
			Status: slackrender.TransitionFailed,
			Error:  "terraform plan exited 1",
		},
	}

	got := Build(e, nil, "", "")

	if got.AlertType != client.EventAlertTypeError {
		t.Errorf("AlertType = %q, want error", got.AlertType)
	}
	if got.Priority != client.EventPriorityNormal {
		t.Errorf("Priority = %q, want normal", got.Priority)
	}
	if got.AggregationKey != "wf-1" {
		t.Errorf("AggregationKey = %q, want wf-1", got.AggregationKey)
	}
	if got.SourceTypeName != "nuon" {
		t.Errorf("SourceTypeName = %q, want nuon", got.SourceTypeName)
	}
	wantTitle := "Nuon: provision on install acme-prod failed"
	if got.Title != wantTitle {
		t.Errorf("Title = %q, want %q", got.Title, wantTitle)
	}
	if !strings.Contains(got.Text, "**Error:** terraform plan exited 1") {
		t.Errorf("Text missing error excerpt:\n%s", got.Text)
	}
	mustContainTags(t, got.Tags,
		"source:nuon",
		"nuon_kind:workflow",
		"nuon_transition:failed",
		"nuon_org_id:org-1",
		"nuon_workflow_id:wf-1",
		"nuon_workflow_type:provision",
		"nuon_install_id:ins-1",
		"nuon_status:failed",
	)
}

func TestBuild_ApprovalRequestUsesWarning(t *testing.T) {
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflowStepApproval,
		Transition: slackrender.TransitionRequested,
		OrgID:      "org-1",
		Workflow: slackrender.WorkflowRef{
			ID:        "wf-1",
			Type:      slackrender.WorkflowTypeProvision,
			OwnerID:   "ins-1",
			OwnerType: slackrender.OwnerTypeInstalls,
			OwnerName: "acme-prod",
		},
		Step: &slackrender.StepRef{
			ID:   "stp-1",
			Name: "terraform-plan",
		},
		Approval: &slackrender.ApprovalRef{
			ID:   "apv-1",
			Type: "manual",
		},
	}
	got := Build(e, nil, "", "")

	if got.AlertType != client.EventAlertTypeWarning {
		t.Errorf("AlertType = %q, want warning", got.AlertType)
	}
	if !strings.Contains(got.Title, "approval for terraform-plan") {
		t.Errorf("Title doesn't mention step name: %q", got.Title)
	}
	if !strings.Contains(got.Title, "awaiting approval") {
		t.Errorf("Title doesn't mention approval state: %q", got.Title)
	}
}

func TestBuild_ParentRefDrivesAggregationKey(t *testing.T) {
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflow,
		Transition: slackrender.TransitionSucceeded,
		Workflow: slackrender.WorkflowRef{
			ID:   "child-wf",
			Type: slackrender.WorkflowTypeActionWorkflowRun,
		},
		Parent: &slackrender.ParentRef{
			WorkflowID: "parent-wf",
		},
		Outcome: &slackrender.Outcome{Status: slackrender.TransitionSucceeded},
	}
	got := Build(e, nil, "", "")
	if got.AggregationKey != "parent-wf" {
		t.Errorf("AggregationKey = %q, want parent-wf (so DD rolls up under parent)", got.AggregationKey)
	}
}

func TestBuild_ExtraTagsMergedAndDeduped(t *testing.T) {
	e := slackrender.Event{
		Kind: slackrender.KindWorkflow,
		Workflow: slackrender.WorkflowRef{
			ID: "wf-1",
		},
	}
	got := Build(e, []string{"env:prod", "customer:acme", "env:prod", "  "}, "", "")
	mustContainTags(t, got.Tags, "env:prod", "customer:acme")
	if countTag(got.Tags, "env:prod") != 1 {
		t.Errorf("env:prod was not deduped: %v", got.Tags)
	}
	// Confirm sorted output (stable order matters for snapshot tests).
	if !slices.IsSorted(got.Tags) {
		t.Errorf("tags are not sorted: %v", got.Tags)
	}
}

func TestBuild_OverridesWin(t *testing.T) {
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflow,
		Transition: slackrender.TransitionSucceeded,
		Workflow:   slackrender.WorkflowRef{ID: "wf-1"},
		Outcome:    &slackrender.Outcome{Status: slackrender.TransitionSucceeded},
	}
	got := Build(e, nil, "warning", "low")
	if got.AlertType != client.EventAlertType("warning") {
		t.Errorf("AlertTypeOverride ignored: %v", got.AlertType)
	}
	if got.Priority != client.EventPriority("low") {
		t.Errorf("PriorityOverride ignored: %v", got.Priority)
	}
}

func TestBuild_StartedTransitionIsLowPriority(t *testing.T) {
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflow,
		Transition: slackrender.TransitionStarted,
		Workflow:   slackrender.WorkflowRef{ID: "wf-1"},
	}
	got := Build(e, nil, "", "")
	if got.Priority != client.EventPriorityLow {
		t.Errorf("Started events should be low priority (informational), got %q", got.Priority)
	}
}

func TestBuild_DateHappenedFromWorkflowCreatedAt(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	e := slackrender.Event{
		Kind:       slackrender.KindWorkflow,
		Transition: slackrender.TransitionFailed,
		Workflow: slackrender.WorkflowRef{
			ID:        "wf-1",
			CreatedAt: start,
		},
		Outcome: &slackrender.Outcome{Status: slackrender.TransitionFailed},
	}
	got := Build(e, nil, "", "")
	if got.DateHappened != start.Unix() {
		t.Errorf("DateHappened = %d, want %d", got.DateHappened, start.Unix())
	}
}

func mustContainTags(t *testing.T, got []string, want ...string) {
	t.Helper()
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing tag %q in %v", w, got)
		}
	}
}

func countTag(tags []string, tag string) int {
	n := 0
	for _, t := range tags {
		if t == tag {
			n++
		}
	}
	return n
}
