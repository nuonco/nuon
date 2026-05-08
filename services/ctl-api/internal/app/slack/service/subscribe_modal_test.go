package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

func TestSubscribeModalStateRoundTrip(t *testing.T) {
	in := subscribeModalState{
		TeamID:      "T123",
		ChannelID:   "C456",
		ChannelName: "ops",
		SlackUserID: "U789",
	}
	enc, err := encodeSubscribeModalState(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	out, err := decodeSubscribeModalState(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: in=%+v out=%+v", in, out)
	}

	// Empty private_metadata is rejected so callers always know whether
	// state was set.
	if _, err := decodeSubscribeModalState(""); err == nil {
		t.Fatal("expected error decoding empty state")
	}
}

func TestBuildSubscribeModalView(t *testing.T) {
	state := subscribeModalState{TeamID: "T1", ChannelID: "C1", ChannelName: "general", SlackUserID: "U1"}
	links := []app.SlackOrgLink{
		{ID: "l1", TeamID: "T1", OrgID: "org1", Status: app.SlackOrgLinkStatusVerified, Org: app.Org{Name: "Acme"}},
		{ID: "l2", TeamID: "T1", OrgID: "org2", Status: app.SlackOrgLinkStatusVerified, Org: app.Org{Name: "Beta"}},
	}

	t.Run("default match=All renders org+match+notif, no kind/predicate/entities/labels blocks", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		mustContainID(t, ids, subscribeOrgBlockID)
		mustContainID(t, ids, subscribeMatchBlockID)
		mustContainID(t, ids, subscribeNotifBlockID)
		mustNotContainID(t, ids, subscribeKindBlockID)
		mustNotContainID(t, ids, subscribePredicateBlockID)
		mustNotContainID(t, ids, subscribeEntitiesBlockID)
		mustNotContainID(t, ids, subscribeLabelsBlockID)
		// No per-resource blocks render until notif=specific.
		for _, kind := range interests.AllResources {
			mustNotContainID(t, ids, subscribeResourceOptsBlockID(kind))
			mustNotContainID(t, ids, subscribeResourceCategoriesBlockID(kind))
			mustNotContainID(t, ids, subscribeResourceLifecycleBlockID(kind))
			mustNotContainID(t, ids, subscribeResourceSubOpsBlockID(kind))
		}

		// Notif options for the new modal are All / Specific only —
		// the legacy Mute option is gone with the install pin.
		notifOpts := readRadioOptions(t, view, subscribeNotifBlockID, subscribeNotifActionID)
		for _, o := range notifOpts {
			m, _ := o.(map[string]any)
			if v, _ := m["value"].(string); v != notifOptionAll && v != notifOptionSpecific {
				t.Fatalf("unexpected notif option %q", v)
			}
		}

		// private_metadata round-trips.
		pm, _ := view["private_metadata"].(string)
		if !strings.Contains(pm, "T1") || !strings.Contains(pm, "C1") {
			t.Fatalf("private_metadata missing state: %q", pm)
		}
	})

	t.Run("match=Specific defaults render kind+predicate=Any, no entities/labels", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{Match: matchOptionSpecific}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		mustContainID(t, ids, subscribeMatchBlockID)
		mustContainID(t, ids, subscribeKindBlockID)
		mustContainID(t, ids, subscribePredicateBlockID)
		mustNotContainID(t, ids, subscribeEntitiesBlockID)
		mustNotContainID(t, ids, subscribeLabelsBlockID)

		// Default kind is Installs.
		kindBlock := findBlockByID(t, view, subscribeKindBlockID)
		init := kindBlock["element"].(map[string]any)["initial_option"].(map[string]any)
		if init["value"] != kindOptionInstalls {
			t.Fatalf("default kind: got %v want %q", init["value"], kindOptionInstalls)
		}
	})

	for _, tc := range []struct {
		name           string
		render         subscribeModalRenderState
		wantBlockID    string
		wantNoBlockIDs []string
	}{
		{
			name: "match=Specific kind=Installs predicate=Specific renders entity picker (installs action_id)",
			render: subscribeModalRenderState{
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindInstalls,
				Predicate:  predicateOptionSpecific,
			},
			wantBlockID:    subscribeEntitiesBlockID,
			wantNoBlockIDs: []string{subscribeLabelsBlockID},
		},
		{
			name: "match=Specific kind=Components predicate=Specific with app picked renders entity picker (components action_id)",
			render: subscribeModalRenderState{
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindComponents,
				Predicate:  predicateOptionSpecific,
				AppID:      "app1",
				AppName:    "demo",
			},
			wantBlockID:    subscribeEntitiesBlockID,
			wantNoBlockIDs: []string{subscribeLabelsBlockID},
		},
		{
			name: "match=Specific kind=Actions predicate=Specific with app picked renders entity picker (actions action_id)",
			render: subscribeModalRenderState{
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindActions,
				Predicate:  predicateOptionSpecific,
				AppID:      "app1",
				AppName:    "demo",
			},
			wantBlockID:    subscribeEntitiesBlockID,
			wantNoBlockIDs: []string{subscribeLabelsBlockID},
		},
		{
			name: "match=Specific kind=Installs predicate=Labels renders labels textinput",
			render: subscribeModalRenderState{
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindInstalls,
				Predicate:  predicateOptionLabels,
				LabelsRaw:  "env=prod",
			},
			wantBlockID:    subscribeLabelsBlockID,
			wantNoBlockIDs: []string{subscribeEntitiesBlockID},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			view, err := buildSubscribeModalView(state, links, tc.render, nil)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			ids := blockIDs(view)
			mustContainID(t, ids, tc.wantBlockID)
			for _, id := range tc.wantNoBlockIDs {
				mustNotContainID(t, ids, id)
			}
			if tc.render.Predicate == predicateOptionSpecific {
				block := findBlockByID(t, view, subscribeEntitiesBlockID)
				el := block["element"].(map[string]any)
				gotActionID, _ := el["action_id"].(string)
				wantActionID := subscribeEntitiesActionIDForKind(tc.render.TargetKind)
				if gotActionID != wantActionID {
					t.Fatalf("entity picker action_id: got %q want %q", gotActionID, wantActionID)
				}
			}
			if tc.render.Predicate == predicateOptionLabels {
				block := findBlockByID(t, view, subscribeLabelsBlockID)
				el := block["element"].(map[string]any)
				if el["initial_value"] != tc.render.LabelsRaw {
					t.Fatalf("labels initial_value: got %v want %q", el["initial_value"], tc.render.LabelsRaw)
				}
			}
		})
	}

	t.Run("preview block renders only when predicate≠Any and preview!=nil", func(t *testing.T) {
		render := subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionLabels,
			LabelsRaw:  "env=prod",
		}
		preview := &subscribePreview{Kind: labels.TargetKindInstalls, Count: 3, Names: []string{"a", "b", "c"}}
		view, err := buildSubscribeModalView(state, links, render, preview)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		// The context block has no block_id so we have to find it by
		// type. Verify by scanning for the mrkdwn body.
		blocks, _ := view["blocks"].([]any)
		found := false
		for _, b := range blocks {
			m, _ := b.(map[string]any)
			if m["type"] != "context" {
				continue
			}
			els, _ := m["elements"].([]any)
			if len(els) == 0 {
				continue
			}
			el, _ := els[0].(map[string]any)
			text, _ := el["text"].(string)
			if strings.Contains(text, "Currently matches *3* installs") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected preview context block in view")
		}
	})

	t.Run("preview suppressed when predicate=Any", func(t *testing.T) {
		render := subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionAny,
		}
		preview := &subscribePreview{Kind: labels.TargetKindInstalls, Count: 5, Names: []string{"x"}}
		view, err := buildSubscribeModalView(state, links, render, preview)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		blocks, _ := view["blocks"].([]any)
		for _, b := range blocks {
			m, _ := b.(map[string]any)
			if m["type"] == "context" {
				t.Fatal("preview context block should be suppressed for Any")
			}
		}
	})

	t.Run("notif=specific renders categories + lifecycle + subops per enabled resource, seeded from Default()", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{Notif: notifOptionSpecific}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		for _, kind := range interests.AllResources {
			mustContainID(t, ids, subscribeResourceOptsBlockID(kind))
		}
		enabledByDefault := []interests.ResourceKind{
			interests.ResourceInstalls,
			interests.ResourceComponents,
			interests.ResourceSandboxes,
			interests.ResourceInstallConfigurations,
		}
		for _, kind := range enabledByDefault {
			mustContainID(t, ids, subscribeResourceCategoriesBlockID(kind))
			mustContainID(t, ids, subscribeResourceLifecycleBlockID(kind))
			mustContainID(t, ids, subscribeResourceSubOpsBlockID(kind))
		}
		for _, kind := range []interests.ResourceKind{interests.ResourceRunners, interests.ResourceActions} {
			mustNotContainID(t, ids, subscribeResourceCategoriesBlockID(kind))
			mustNotContainID(t, ids, subscribeResourceLifecycleBlockID(kind))
			mustNotContainID(t, ids, subscribeResourceSubOpsBlockID(kind))
		}
	})

	t.Run("empty links rejected", func(t *testing.T) {
		if _, err := buildSubscribeModalView(state, nil, subscribeModalRenderState{}, nil); err == nil {
			t.Fatal("expected error with empty links")
		}
	})
}

func TestReadSubscribeRenderStateFromPayload_MatchAndPredicate(t *testing.T) {
	cases := []struct {
		name      string
		stateJSON string
		want      subscribeModalRenderState
	}{
		{
			name: "match=All",
			stateJSON: `{
				"nuon_subscribe_org_block": {"nuon_subscribe_org": {"type":"static_select","selected_option": {"value":"l1"}}},
				"nuon_subscribe_match_block": {"nuon_subscribe_match": {"type":"radio_buttons","selected_option":{"value":"all"}}},
				"nuon_subscribe_notif_block": {"nuon_subscribe_notif": {"type":"radio_buttons","selected_option":{"value":"all"}}}
			}`,
			want: subscribeModalRenderState{
				OrgLinkID:  "l1",
				Match:      matchOptionAll,
				TargetKind: labels.TargetKindInstalls,
				Notif:      notifOptionAll,
			},
		},
		{
			name: "match=Specific kind=Components predicate=Specific reads entity ids from components action_id",
			stateJSON: `{
				"nuon_subscribe_org_block": {"nuon_subscribe_org": {"type":"static_select","selected_option": {"value":"l1"}}},
				"nuon_subscribe_match_block": {"nuon_subscribe_match": {"type":"radio_buttons","selected_option":{"value":"specific"}}},
				"nuon_subscribe_kind_block": {"nuon_subscribe_kind": {"type":"static_select","selected_option":{"value":"components"}}},
				"nuon_subscribe_predicate_block": {"nuon_subscribe_predicate": {"type":"radio_buttons","selected_option":{"value":"specific"}}},
				"nuon_subscribe_entities_block": {"nuon_subscribe_entities_components": {
					"type":"multi_external_select",
					"selected_options": [
						{"text":{"type":"plain_text","text":"web"},"value":"comp_a"},
						{"text":{"type":"plain_text","text":"db"},"value":"comp_b"}
					]
				}},
				"nuon_subscribe_notif_block": {"nuon_subscribe_notif": {"type":"radio_buttons","selected_option":{"value":"all"}}}
			}`,
			want: subscribeModalRenderState{
				OrgLinkID:   "l1",
				Match:       matchOptionSpecific,
				TargetKind:  labels.TargetKindComponents,
				Predicate:   predicateOptionSpecific,
				EntityIDs:   []string{"comp_a", "comp_b"},
				EntityNames: []string{"web", "db"},
				Notif:       notifOptionAll,
			},
		},
		{
			name: "match=Specific kind=Installs predicate=Labels reads raw text",
			stateJSON: `{
				"nuon_subscribe_org_block": {"nuon_subscribe_org": {"type":"static_select","selected_option": {"value":"l1"}}},
				"nuon_subscribe_match_block": {"nuon_subscribe_match": {"type":"radio_buttons","selected_option":{"value":"specific"}}},
				"nuon_subscribe_kind_block": {"nuon_subscribe_kind": {"type":"static_select","selected_option":{"value":"installs"}}},
				"nuon_subscribe_predicate_block": {"nuon_subscribe_predicate": {"type":"radio_buttons","selected_option":{"value":"labels"}}},
				"nuon_subscribe_labels_block": {"nuon_subscribe_labels": {"type":"plain_text_input","value":"env=prod, tier=critical"}},
				"nuon_subscribe_notif_block": {"nuon_subscribe_notif": {"type":"radio_buttons","selected_option":{"value":"all"}}}
			}`,
			want: subscribeModalRenderState{
				OrgLinkID:  "l1",
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindInstalls,
				Predicate:  predicateOptionLabels,
				LabelsRaw:  "env=prod, tier=critical",
				Notif:      notifOptionAll,
			},
		},
		{
			name: "kind switch drops stale install entries when current kind=Components",
			stateJSON: `{
				"nuon_subscribe_org_block": {"nuon_subscribe_org": {"type":"static_select","selected_option": {"value":"l1"}}},
				"nuon_subscribe_match_block": {"nuon_subscribe_match": {"type":"radio_buttons","selected_option":{"value":"specific"}}},
				"nuon_subscribe_kind_block": {"nuon_subscribe_kind": {"type":"static_select","selected_option":{"value":"components"}}},
				"nuon_subscribe_predicate_block": {"nuon_subscribe_predicate": {"type":"radio_buttons","selected_option":{"value":"specific"}}},
				"nuon_subscribe_entities_block": {"nuon_subscribe_entities_installs": {
					"type":"multi_external_select",
					"selected_options": [{"text":{"type":"plain_text","text":"old"},"value":"in_old"}]
				}},
				"nuon_subscribe_notif_block": {"nuon_subscribe_notif": {"type":"radio_buttons","selected_option":{"value":"all"}}}
			}`,
			want: subscribeModalRenderState{
				OrgLinkID:  "l1",
				Match:      matchOptionSpecific,
				TargetKind: labels.TargetKindComponents,
				Predicate:  predicateOptionSpecific,
				Notif:      notifOptionAll,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawJSON := `{"type":"view_submission","team":{"id":"T1"},"user":{"id":"U1"},"view":{"id":"V1","callback_id":"nuon_subscribe_modal","private_metadata":"{\"team_id\":\"T1\",\"channel_id\":\"C1\",\"channel_name\":\"ops\",\"slack_user_id\":\"U1\"}","state":{"values":` + tc.stateJSON + `}}}`
			var p slackInteractionPayload
			if err := json.Unmarshal([]byte(rawJSON), &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			got := readSubscribeRenderStateFromPayload(p)
			// Resources is computed separately and varies; don't compare.
			got.Resources = nil
			if !renderStateEq(got, tc.want) {
				t.Fatalf("render mismatch:\n got: %+v\nwant: %+v", got, tc.want)
			}
		})
	}
}

func TestBuildSubscriptionMatchFromRender(t *testing.T) {
	t.Run("Match=All -> nil match, no error", func(t *testing.T) {
		m, _, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{Match: matchOptionAll})
		if msg != "" {
			t.Fatalf("unexpected error: %q", msg)
		}
		if m != nil {
			t.Fatalf("expected nil match, got %+v", m)
		}
	})

	t.Run("Predicate=Any -> empty TargetMatch on chosen kind", func(t *testing.T) {
		m, _, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindActions,
			Predicate:  predicateOptionAny,
		})
		if msg != "" {
			t.Fatalf("unexpected error: %q", msg)
		}
		if m == nil || m.Actions == nil {
			t.Fatalf("expected Actions populated, got %+v", m)
		}
		if len(m.Actions.IDs) != 0 || m.Actions.Selector != nil {
			t.Fatalf("expected empty TargetMatch, got %+v", m.Actions)
		}
	})

	t.Run("Predicate=Specific empty entities -> error keyed on entities block", func(t *testing.T) {
		_, block, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionSpecific,
		})
		if block != subscribeEntitiesBlockID {
			t.Fatalf("expected error block %q, got %q", subscribeEntitiesBlockID, block)
		}
		if msg == "" {
			t.Fatal("expected error message")
		}
	})

	t.Run("Predicate=Labels empty raw -> error keyed on labels block", func(t *testing.T) {
		_, block, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionLabels,
			LabelsRaw:  "   ",
		})
		if block != subscribeLabelsBlockID {
			t.Fatalf("expected error block %q, got %q", subscribeLabelsBlockID, block)
		}
		if msg == "" {
			t.Fatal("expected error message")
		}
	})

	t.Run("Predicate=Labels valid raw -> Selector populated", func(t *testing.T) {
		m, _, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindComponents,
			Predicate:  predicateOptionLabels,
			LabelsRaw:  "env=prod, owner=*",
		})
		if msg != "" {
			t.Fatalf("unexpected error: %q", msg)
		}
		if m == nil || m.Components == nil || m.Components.Selector == nil {
			t.Fatalf("expected Components.Selector populated, got %+v", m)
		}
		if got := m.Components.Selector.MatchLabels["env"]; got != "prod" {
			t.Fatalf("env: got %q want %q", got, "prod")
		}
		if got := m.Components.Selector.MatchLabels["owner"]; got != "*" {
			t.Fatalf("owner: got %q want %q", got, "*")
		}
	})

	t.Run("Predicate=Specific with ids -> IDs populated", func(t *testing.T) {
		m, _, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionSpecific,
			EntityIDs:  []string{"in_a", "in_b"},
		})
		if msg != "" {
			t.Fatalf("unexpected error: %q", msg)
		}
		if m == nil || m.Installs == nil {
			t.Fatalf("expected Installs populated, got %+v", m)
		}
		if !stringSliceEq(m.Installs.IDs, []string{"in_a", "in_b"}) {
			t.Fatalf("installs ids: got %v", m.Installs.IDs)
		}
	})

	t.Run("Predicate=Specific kind=Components without app -> error keyed on app block", func(t *testing.T) {
		_, block, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindComponents,
			Predicate:  predicateOptionSpecific,
			EntityIDs:  []string{"comp_a"},
		})
		if block != subscribeAppBlockID {
			t.Fatalf("expected error block %q, got %q", subscribeAppBlockID, block)
		}
		if msg == "" {
			t.Fatal("expected error message")
		}
	})

	t.Run("Predicate=Specific kind=Actions with app + ids -> Actions populated", func(t *testing.T) {
		m, _, msg := buildSubscriptionMatchFromRender(subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindActions,
			Predicate:  predicateOptionSpecific,
			AppID:      "app1",
			EntityIDs:  []string{"act_a"},
		})
		if msg != "" {
			t.Fatalf("unexpected error: %q", msg)
		}
		if m == nil || m.Actions == nil || !stringSliceEq(m.Actions.IDs, []string{"act_a"}) {
			t.Fatalf("expected Actions populated, got %+v", m)
		}
	})
}

func TestBuildSubscribeModalView_AppPickerGate(t *testing.T) {
	state := subscribeModalState{TeamID: "T1", ChannelID: "C1", ChannelName: "general", SlackUserID: "U1"}
	links := []app.SlackOrgLink{
		{ID: "l1", TeamID: "T1", OrgID: "org1", Status: app.SlackOrgLinkStatusVerified, Org: app.Org{Name: "Acme"}},
	}

	t.Run("kind=Components predicate=Specific without app picked: app block rendered, entities block hidden", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindComponents,
			Predicate:  predicateOptionSpecific,
		}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		mustContainID(t, ids, subscribeAppBlockID)
		mustNotContainID(t, ids, subscribeEntitiesBlockID)
	})

	t.Run("kind=Installs predicate=Specific: app block NOT rendered", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindInstalls,
			Predicate:  predicateOptionSpecific,
		}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		mustNotContainID(t, ids, subscribeAppBlockID)
		mustContainID(t, ids, subscribeEntitiesBlockID)
	})

	t.Run("kind=Components predicate=Labels: app block NOT rendered (labels span apps)", func(t *testing.T) {
		view, err := buildSubscribeModalView(state, links, subscribeModalRenderState{
			Match:      matchOptionSpecific,
			TargetKind: labels.TargetKindComponents,
			Predicate:  predicateOptionLabels,
			LabelsRaw:  "env=prod",
		}, nil)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		ids := blockIDs(view)
		mustNotContainID(t, ids, subscribeAppBlockID)
		mustContainID(t, ids, subscribeLabelsBlockID)
	})
}

func TestLabelsToQueryString(t *testing.T) {
	got := labelsToQueryString(labels.Labels{"env": "prod", "owner": "*", "tier": "critical"})
	// Keys sorted alphabetically.
	want := "env=prod, owner=*, tier=critical"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := labelsToQueryString(nil); got != "" {
		t.Fatalf("empty: got %q", got)
	}
}

func TestRoundTripBuildSpecificEventsInterests(t *testing.T) {
	t.Run("disabled resources are dropped", func(t *testing.T) {
		i := buildSpecificEventsInterests(map[interests.ResourceKind]subscribeResourceCfg{
			interests.ResourceInstalls:   {Enabled: false, Approvals: true, Outcome: outcomeOptionAll},
			interests.ResourceComponents: {Enabled: true, Outcome: outcomeOptionCompletion},
		})
		if _, ok := i.Resources[interests.ResourceInstalls]; ok {
			t.Fatal("disabled resources must be dropped")
		}
		if _, ok := i.Resources[interests.ResourceComponents]; !ok {
			t.Fatal("enabled resources must be present")
		}
	})

	t.Run("approvals collapse mirrors to both ApprovalRequests + ApprovalResponses", func(t *testing.T) {
		i := buildSpecificEventsInterests(map[interests.ResourceKind]subscribeResourceCfg{
			interests.ResourceInstalls: {Enabled: true, Approvals: true, Outcome: outcomeOptionCompletion},
		})
		cfg := i.Resources[interests.ResourceInstalls]
		if !cfg.ApprovalRequests || !cfg.ApprovalResponses {
			t.Fatalf("expected both approval flags true, got %+v", cfg)
		}
	})

	t.Run("ops are filtered to interests.SubOps[kind] vocabulary", func(t *testing.T) {
		i := buildSpecificEventsInterests(map[interests.ResourceKind]subscribeResourceCfg{
			interests.ResourceInstalls: {
				Enabled: true,
				Ops:     []string{"provision", "deploy", "bogus"},
				Outcome: outcomeOptionAll,
			},
		})
		cfg := i.Resources[interests.ResourceInstalls]
		if !stringSliceEq(cfg.Ops, []string{"provision"}) {
			t.Fatalf("expected only provision to survive filter, got %v", cfg.Ops)
		}
	})

	t.Run("drift only persists for resources that support it", func(t *testing.T) {
		i := buildSpecificEventsInterests(map[interests.ResourceKind]subscribeResourceCfg{
			interests.ResourceComponents: {Enabled: true, Drift: true, Outcome: outcomeOptionAll},
			interests.ResourceInstalls:   {Enabled: true, Drift: true, Outcome: outcomeOptionAll},
		})
		if !i.Resources[interests.ResourceComponents].DriftDetected {
			t.Fatal("components should keep drift_detected=true")
		}
		if i.Resources[interests.ResourceInstalls].DriftDetected {
			t.Fatal("installs do not support drift_detected; flag must be dropped")
		}
	})
}

func TestModalErr(t *testing.T) {
	got := modalErr(subscribeEntitiesBlockID, "Pick something.")
	if got["response_action"] != "errors" {
		t.Fatalf("response_action: got %v", got["response_action"])
	}
	errs, ok := got["errors"].(map[string]string)
	if !ok {
		t.Fatalf("errors: bad type %T", got["errors"])
	}
	if errs[subscribeEntitiesBlockID] != "Pick something." {
		t.Fatalf("error message mismatch: %v", errs)
	}
}

func TestTruncatePlainText(t *testing.T) {
	if got := truncatePlainText("short", 10); got != "short" {
		t.Fatalf("short: got %q", got)
	}
	// "…" is U+2026 = 3 UTF-8 bytes, so a max=5 truncation reserves 3
	// bytes for the ellipsis and 2 bytes for content → "01…" (5 bytes).
	if got := truncatePlainText("0123456789", 5); got != "01…" {
		t.Fatalf("long: got %q (len=%d)", got, len(got))
	}
	// max smaller than the ellipsis falls through to a hard rune-safe
	// truncation; ASCII input lets us assert exact bytes.
	if got := truncatePlainText("abcdef", 1); got != "a" {
		t.Fatalf("max=1: got %q", got)
	}
	// 75-byte cap is the Slack contract; assert the post-condition the
	// block_suggestion handlers depend on.
	if got := truncatePlainText(strings.Repeat("x", 200), 75); len(got) > 75 {
		t.Fatalf("max=75: got len=%d (must be <=75)", len(got))
	}
}

// --- helpers ---

func blockIDs(view map[string]any) []string {
	blocks, _ := view["blocks"].([]any)
	out := make([]string, 0, len(blocks))
	for _, b := range blocks {
		m, _ := b.(map[string]any)
		if id, ok := m["block_id"].(string); ok {
			out = append(out, id)
		}
	}
	return out
}

func mustContainID(t *testing.T, ids []string, want string) {
	t.Helper()
	for _, id := range ids {
		if id == want {
			return
		}
	}
	t.Fatalf("missing block id %q in %v", want, ids)
}

func mustNotContainID(t *testing.T, ids []string, unwanted string) {
	t.Helper()
	for _, id := range ids {
		if id == unwanted {
			t.Fatalf("unexpected block id %q in %v", unwanted, ids)
		}
	}
}

func findBlockByID(t *testing.T, view map[string]any, id string) map[string]any {
	t.Helper()
	blocks, _ := view["blocks"].([]any)
	for _, b := range blocks {
		m, _ := b.(map[string]any)
		if m["block_id"] == id {
			return m
		}
	}
	t.Fatalf("block %q not found", id)
	return nil
}

func readRadioOptions(t *testing.T, view map[string]any, blockID, _ string) []any {
	t.Helper()
	block := findBlockByID(t, view, blockID)
	element, _ := block["element"].(map[string]any)
	opts, _ := element["options"].([]any)
	return opts
}

func stringSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func renderStateEq(a, b subscribeModalRenderState) bool {
	if a.OrgLinkID != b.OrgLinkID || a.Match != b.Match || a.TargetKind != b.TargetKind ||
		a.Predicate != b.Predicate || a.LabelsRaw != b.LabelsRaw || a.Notif != b.Notif {
		return false
	}
	if !stringSliceEq(a.EntityIDs, b.EntityIDs) || !stringSliceEq(a.EntityNames, b.EntityNames) {
		return false
	}
	return true
}
