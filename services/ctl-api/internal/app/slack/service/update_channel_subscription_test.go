package service

import (
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/pkg/labels"
)

// TestUpdateChannelSubscriptionRequestUnmarshal verifies the sentinel
// detection that lets the handler distinguish "leave match unchanged"
// (omit the key) from "make this org-wide" (explicit JSON null) — without
// it, the handler couldn't ever clear a per-resource scope back to org-wide.
func TestUpdateChannelSubscriptionRequestUnmarshal(t *testing.T) {
	cases := []struct {
		name         string
		jsonBody     string
		wantSet      bool
		wantNilMatch bool
	}{
		{
			name:         "match key omitted leaves MatchSet=false",
			jsonBody:     `{"channel_name":"deploys"}`,
			wantSet:      false,
			wantNilMatch: true,
		},
		{
			name:         "explicit null match sets MatchSet=true and Match=nil",
			jsonBody:     `{"match":null}`,
			wantSet:      true,
			wantNilMatch: true,
		},
		{
			name:         "populated match sets MatchSet=true with shape",
			jsonBody:     `{"match":{"installs":{"ids":["i_a"]}}}`,
			wantSet:      true,
			wantNilMatch: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req UpdateChannelSubscriptionRequest
			if err := json.Unmarshal([]byte(tc.jsonBody), &req); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if req.MatchSet != tc.wantSet {
				t.Errorf("MatchSet = %v, want %v", req.MatchSet, tc.wantSet)
			}
			if (req.Match == nil) != tc.wantNilMatch {
				t.Errorf("Match nil-ness = %v, want nil=%v (got %+v)", req.Match == nil, tc.wantNilMatch, req.Match)
			}
		})
	}
}

// TestUpdateChannelSubscriptionRequestRoundTrip covers the typical
// dashboard payload — the SubscriptionMatch shape decodes cleanly into
// the typed struct so handler code can read sub.Match.Installs.IDs
// without re-decoding raw JSON.
func TestUpdateChannelSubscriptionRequestRoundTrip(t *testing.T) {
	body := `{"match":{"installs":{"ids":["i_a","i_b"]}},"interests":{"all_events":true}}`
	var req UpdateChannelSubscriptionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !req.MatchSet {
		t.Fatalf("expected MatchSet=true")
	}
	if req.Match == nil || req.Match.Installs == nil {
		t.Fatalf("expected Match.Installs populated, got %+v", req.Match)
	}
	got := req.Match.Installs.IDs
	want := []string{"i_a", "i_b"}
	if len(got) != len(want) {
		t.Fatalf("ids len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if req.Interests == nil || !req.Interests.AllEvents {
		t.Errorf("expected interests.AllEvents=true, got %+v", req.Interests)
	}
	// Touch labels package so import isn't unused if the build trims it
	// during test compilation in oddly tagged builds.
	_ = labels.SubscriptionMatch{}
}
