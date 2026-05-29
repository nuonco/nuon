package service

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestBuildMonitorQuery_AllSupportedCombos pins the exact query strings
// the dashboard's one-click action emits. Drift here will silently break
// users' alerts (DD will report "no events match" and never fire), so
// the queries are asserted as full literal strings rather than via
// regex/contains.
func TestBuildMonitorQuery_AllSupportedCombos(t *testing.T) {
	cases := []struct {
		name       string
		targetType app.DatadogManagedMonitorTargetType
		targetID   string
		installID  string
		preset     app.DatadogManagedMonitorPreset
		want       string
	}{
		{
			name:       "install failure",
			targetType: app.DatadogManagedMonitorTargetTypeInstall,
			targetID:   "inst123",
			preset:     app.DatadogManagedMonitorPresetFailure,
			want:       `events("source:nuon nuon_install_id:inst123 nuon_status:failed").rollup("count").last("5m") >= 1`,
		},
		{
			name:       "component failure",
			targetType: app.DatadogManagedMonitorTargetTypeComponent,
			targetID:   "cmp456",
			preset:     app.DatadogManagedMonitorPresetFailure,
			want:       `events("source:nuon nuon_component_id:cmp456 nuon_status:failed").rollup("count").last("5m") >= 1`,
		},
		{
			name:       "workflow failure",
			targetType: app.DatadogManagedMonitorTargetTypeWorkflow,
			targetID:   "wf789",
			preset:     app.DatadogManagedMonitorPresetFailure,
			want:       `events("source:nuon nuon_workflow_id:wf789 nuon_status:failed").rollup("count").last("5m") >= 1`,
		},
		{
			name:       "install drift",
			targetType: app.DatadogManagedMonitorTargetTypeInstall,
			targetID:   "inst123",
			preset:     app.DatadogManagedMonitorPresetDrift,
			want:       `events("source:nuon nuon_install_id:inst123 nuon_kind:drift").rollup("count").last("5m") >= 1`,
		},
		{
			// Action target = org-level action_workflows.id ANDed with
			// the install scope. The dashboard's one-click flow on an
			// install's action page always passes installID so the
			// resulting monitor only fires for that install's runs of
			// the action — mirrors install/component/workflow per-
			// resource targeting semantics.
			name:       "action failure (install-scoped)",
			targetType: app.DatadogManagedMonitorTargetTypeAction,
			targetID:   "actwfl999",
			installID:  "inst123",
			preset:     app.DatadogManagedMonitorPresetFailure,
			want:       `events("source:nuon nuon_action_id:actwfl999 nuon_install_id:inst123 nuon_status:failed").rollup("count").last("5m") >= 1`,
		},
		{
			// install_id supplied for a non-action target narrows the
			// alert to one install — useful when one customer's
			// install is flaky on a shared component but org-wide
			// noise should be suppressed.
			name:       "component failure scoped to install",
			targetType: app.DatadogManagedMonitorTargetTypeComponent,
			targetID:   "cmp456",
			installID:  "inst123",
			preset:     app.DatadogManagedMonitorPresetFailure,
			want:       `events("source:nuon nuon_component_id:cmp456 nuon_install_id:inst123 nuon_status:failed").rollup("count").last("5m") >= 1`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildMonitorQuery(tc.targetType, tc.targetID, tc.installID, tc.preset)
			if err != nil {
				t.Fatalf("buildMonitorQuery returned error: %v", err)
			}
			if got != tc.want {
				t.Errorf("query mismatch\n  got:  %s\n  want: %s", got, tc.want)
			}
		})
	}
}

// TestBuildMonitorQuery_ActionRequiresInstallID guards the v1 contract:
// the one-click action button must always send install_id. An org-wide
// action alert would silently match every install's invocations, which
// reads worse than an explicit error and forces callers to opt in.
func TestBuildMonitorQuery_ActionRequiresInstallID(t *testing.T) {
	_, err := buildMonitorQuery(app.DatadogManagedMonitorTargetTypeAction, "actwfl1", "", app.DatadogManagedMonitorPresetFailure)
	if err == nil {
		t.Fatal("expected error for action target without install_id, got nil")
	}
	if !strings.Contains(err.Error(), "install_id is required") {
		t.Errorf("unexpected error text: %v", err)
	}
}

// TestBuildMonitorQuery_EmptyTargetID rejects creation paths that lose
// the target ID — without it we'd emit a query that matches every Nuon
// event of the chosen status, paging on totally unrelated workloads.
func TestBuildMonitorQuery_EmptyTargetID(t *testing.T) {
	_, err := buildMonitorQuery(app.DatadogManagedMonitorTargetTypeInstall, "", "", app.DatadogManagedMonitorPresetFailure)
	if err == nil {
		t.Fatal("expected error for empty target_id, got nil")
	}
}

// TestBuildMonitorRequest_NameAndMessage covers the user-visible strings
// the one-click action plants into DD. Two invariants worth pinning:
//
//  1. The name carries the "[Nuon]" prefix so ops can identify managed
//     monitors in DD's UI without checking tags.
//  2. The message places @-handles last so DD's notification fan-out
//     picks them up (DD only routes @-mentions from the tail of the
//     message body).
func TestBuildMonitorRequest_NameAndMessage(t *testing.T) {
	req, err := buildMonitorRequest(
		app.DatadogManagedMonitorTargetTypeInstall,
		"inst123",
		"",
		app.DatadogManagedMonitorPresetFailure,
		[]string{"@pagerduty-prod", "@slack-oncall"},
		"acme-prod",
		"https://app.nuon.co",
		"org456",
	)
	if err != nil {
		t.Fatalf("buildMonitorRequest returned error: %v", err)
	}

	if !strings.HasPrefix(req.Name, "[Nuon] ") {
		t.Errorf("monitor name should start with \"[Nuon] \", got %q", req.Name)
	}
	if !strings.Contains(req.Name, "acme-prod") {
		t.Errorf("monitor name should include display name, got %q", req.Name)
	}

	// Deep link is the whole point of this body — without it the
	// Slack/PagerDuty alert has no clickable path back to Nuon.
	wantLink := "https://app.nuon.co/org456/installs/inst123"
	if !strings.Contains(req.Message, wantLink) {
		t.Errorf("monitor message must include deep link %q, got %q", wantLink, req.Message)
	}

	// Handles must appear after the deep links so DD's fan-out picks
	// them up — DD only routes @-mentions from the tail of the body.
	linkIdx := strings.Index(req.Message, wantLink)
	handleIdx := strings.Index(req.Message, "@pagerduty-prod")
	if linkIdx == -1 || handleIdx == -1 || handleIdx < linkIdx {
		t.Errorf("notify handles must appear after deep links, got %q", req.Message)
	}

	wantTags := map[string]bool{
		"source:nuon":              false,
		"nuon_managed:true":        false,
		"nuon_target_type:install": false,
		"nuon_target_id:inst123":   false,
		"nuon_preset:failure":      false,
	}
	for _, tag := range req.Tags {
		if _, ok := wantTags[tag]; ok {
			wantTags[tag] = true
		}
	}
	for tag, found := range wantTags {
		if !found {
			t.Errorf("expected tag %q on managed monitor request, missing", tag)
		}
	}
}

// TestBuildMonitorLinks covers the deep-link generation for every
// target type. The dashboard URL shape is part of the Nuon ↔ DD
// contract — drift here means alerts in Slack / PagerDuty / email
// stop being clickable.
func TestBuildMonitorLinks(t *testing.T) {
	const (
		appURL = "https://app.nuon.co"
		orgID  = "org456"
	)
	cases := []struct {
		name       string
		targetType app.DatadogManagedMonitorTargetType
		targetID   string
		installID  string
		want       []string
	}{
		{
			name:       "install always builds an install link",
			targetType: app.DatadogManagedMonitorTargetTypeInstall,
			targetID:   "inst123",
			want: []string{
				"[Open install in Nuon](https://app.nuon.co/org456/installs/inst123)",
				"[Org dashboard](https://app.nuon.co/org456)",
			},
		},
		{
			name:       "action with install scope deep-links to the action page",
			targetType: app.DatadogManagedMonitorTargetTypeAction,
			targetID:   "actwfl999",
			installID:  "inst123",
			want: []string{
				"[Open action in Nuon](https://app.nuon.co/org456/installs/inst123/actions/actwfl999)",
				"[Open install in Nuon](https://app.nuon.co/org456/installs/inst123)",
				"[Org dashboard](https://app.nuon.co/org456)",
			},
		},
		{
			name:       "component with install scope deep-links to the component page",
			targetType: app.DatadogManagedMonitorTargetTypeComponent,
			targetID:   "cmp456",
			installID:  "inst123",
			want: []string{
				"[Open component in Nuon](https://app.nuon.co/org456/installs/inst123/components/cmp456)",
				"[Open install in Nuon](https://app.nuon.co/org456/installs/inst123)",
				"[Org dashboard](https://app.nuon.co/org456)",
			},
		},
		{
			name:       "workflow without install scope falls back to org dashboard only",
			targetType: app.DatadogManagedMonitorTargetTypeWorkflow,
			targetID:   "wf789",
			want: []string{
				"[Org dashboard](https://app.nuon.co/org456)",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildMonitorLinks(tc.targetType, tc.targetID, tc.installID, appURL, orgID)
			if len(got) != len(tc.want) {
				t.Fatalf("link count mismatch\n  got:  %v\n  want: %v", got, tc.want)
			}
			for i, line := range got {
				if line != tc.want[i] {
					t.Errorf("line %d mismatch\n  got:  %s\n  want: %s", i, line, tc.want[i])
				}
			}
		})
	}
}

// TestBuildMonitorLinks_NoConfigReturnsNil ensures the message stays
// link-free in dev/CI where AppURL or the org context is missing,
// rather than emitting broken markdown.
func TestBuildMonitorLinks_NoConfigReturnsNil(t *testing.T) {
	if got := buildMonitorLinks(app.DatadogManagedMonitorTargetTypeInstall, "inst1", "", "", "org1"); got != nil {
		t.Errorf("expected nil for empty AppURL, got %v", got)
	}
	if got := buildMonitorLinks(app.DatadogManagedMonitorTargetTypeInstall, "inst1", "", "https://app.nuon.co", ""); got != nil {
		t.Errorf("expected nil for empty orgID, got %v", got)
	}
}

// TestCleanHandles_DropsEmpty makes sure a stray "" in
// DefaultNotifyHandles doesn't render as a dangling " " in the DD
// monitor body.
func TestCleanHandles_DropsEmpty(t *testing.T) {
	got := cleanHandles([]string{"@one", "", "  ", "@two"})
	want := []string{"@one", "@two"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d, want %d (%v)", len(got), len(want), got)
	}
	for i, h := range got {
		if h != want[i] {
			t.Errorf("index %d: got %q want %q", i, h, want[i])
		}
	}
}
