package orgs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ListChannelSubscriptions(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	subs, err := s.api.ListSlackChannelSubscriptions(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(subs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"CHANNEL",
			"WORKSPACE",
			"SCOPE",
			"CREATED AT",
		},
	}

	for _, sub := range subs {
		data = append(data, []string{
			sub.ID,
			channelLabel(sub),
			sub.TeamID,
			describeSubscriptionMatch(sub.Match),
			sub.CreatedAt,
		})
	}

	view.Render(data)
	return nil
}

// CreateChannelSubscription resolves orgLinkID to the org's single linked
// Slack workspace when empty; with multiple links, the caller must
// disambiguate.
func (s *Service) CreateChannelSubscription(
	ctx context.Context,
	channelID, channelName, orgLinkID string,
	subscription SubscriptionFlags,
	asJSON bool,
) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if channelID == "" {
		return view.Error(fmt.Errorf("channel id is required"))
	}

	if orgLinkID == "" {
		links, err := s.api.ListSlackOrgLinks(ctx)
		if err != nil {
			return view.Error(err)
		}
		switch len(links) {
		case 1:
			orgLinkID = links[0].ID
		case 0:
			return view.Error(fmt.Errorf("this org has no linked Slack workspaces — install the Slack app first"))
		default:
			ids := make([]string, 0, len(links))
			for _, l := range links {
				ids = append(ids, fmt.Sprintf("%s (team %s)", l.ID, l.TeamID))
			}
			return view.Error(fmt.Errorf("this org has multiple linked Slack workspaces — pass --org-link-id with one of: %s", strings.Join(ids, ", ")))
		}
	}

	payload, err := resolveSubscription(ctx, s.api, s.cfg.Interactive, subscription)
	if err != nil {
		return view.Error(err)
	}

	sub, err := s.api.CreateSlackChannelSubscription(ctx, &models.ServiceCreateChannelSubscriptionRequest{
		ChannelID:   &channelID,
		ChannelName: channelName,
		OrgLinkID:   &orgLinkID,
		Interests:   payload.Interests,
		Match:       payload.Match,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(sub)
		return nil
	}

	view.Render([][]string{
		{"id", sub.ID},
		{"channel", channelLabel(sub)},
		{"workspace", sub.TeamID},
		{"scope", describeSubscriptionMatch(sub.Match)},
		{"created at", sub.CreatedAt},
	})
	return nil
}

// UpdateChannelSubscription is a partial update: keys omitted from the patch
// leave the stored value untouched, and "match": null resets the scope to
// org-wide. The TUI picker is intentionally not wired here — it writes one
// resource kind at a time and never pre-loads the existing subscription, so
// multi-kind matches would be silently destroyed; JSON is the only lossless
// way to express them.
func (s *Service) UpdateChannelSubscription(
	ctx context.Context,
	subID string,
	subscription SubscriptionFlags,
	asJSON bool,
) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if subID == "" {
		return view.Error(fmt.Errorf("subscription id is required"))
	}

	rawJSON := strings.TrimSpace(subscription.JSON)
	rawFile := strings.TrimSpace(subscription.File)
	if rawJSON != "" && rawFile != "" {
		return view.Error(fmt.Errorf("only one of --subscription-json, --subscription-file may be set"))
	}
	if rawJSON == "" && rawFile == "" {
		return view.Error(fmt.Errorf("nothing to update — pass --subscription-json or --subscription-file (updates are partial: omitted keys are left unchanged)"))
	}

	var raw []byte
	if rawFile != "" {
		b, err := os.ReadFile(rawFile)
		if err != nil {
			return view.Error(fmt.Errorf("read --subscription-file: %w", err))
		}
		raw = b
	} else {
		raw = []byte(rawJSON)
	}

	req, err := parseChannelSubscriptionPatch(raw)
	if err != nil {
		return view.Error(err)
	}

	sub, err := s.api.UpdateSlackChannelSubscription(ctx, subID, req)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(sub)
		return nil
	}

	view.Render([][]string{
		{"id", sub.ID},
		{"channel", channelLabel(sub)},
		{"workspace", sub.TeamID},
		{"scope", describeSubscriptionMatch(sub.Match)},
		{"updated at", sub.UpdatedAt},
	})
	return nil
}

// parseChannelSubscriptionPatch preserves key presence the same way the API
// does: "match": null (make org-wide) is distinct from an omitted match
// (leave unchanged). Interests is passed through as raw JSON so
// unsupported-but-valid server shapes survive the round trip.
func parseChannelSubscriptionPatch(raw []byte) (*models.ServiceUpdateChannelSubscriptionRequest, error) {
	var rawKeys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawKeys); err != nil {
		return nil, fmt.Errorf("parse subscription JSON: %w", err)
	}
	if len(rawKeys) == 0 {
		return nil, fmt.Errorf("subscription JSON is empty — pass interests and/or match")
	}

	req := &models.ServiceUpdateChannelSubscriptionRequest{}
	if raw, ok := rawKeys["interests"]; ok {
		var interests any
		if err := json.Unmarshal(raw, &interests); err != nil {
			return nil, fmt.Errorf("parse interests: %w", err)
		}
		req.Interests = interests
	}

	rawMatch, matchSet := rawKeys["match"]
	if !matchSet {
		return req, nil
	}

	var match *labels.SubscriptionMatch
	if err := json.Unmarshal(rawMatch, &match); err != nil {
		return nil, fmt.Errorf("parse match: %w", err)
	}
	if match == nil {
		// "match": null — reset to org-wide. The SDK model's omitempty would
		// drop a plain nil, so serialize the null explicitly.
		req.Match = json.RawMessage("null")
		return req, nil
	}
	if err := match.Validate(); err != nil {
		return nil, fmt.Errorf("invalid match: %w", err)
	}
	req.Match = match
	return req, nil
}

func (s *Service) DeleteChannelSubscription(ctx context.Context, subID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	if subID == "" {
		return ui.PrintError(fmt.Errorf("subscription id is required"))
	}

	if asJSON {
		err := s.api.DeleteSlackChannelSubscription(ctx, subID)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		type response struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		ui.PrintJSON(response{
			ID:      subID,
			Deleted: true,
		})
		return nil
	}

	view := ui.NewDeleteView("channel subscription", subID, s.cfg.Interactive)
	view.Start()
	if err := s.api.DeleteSlackChannelSubscription(ctx, subID); err != nil {
		return view.Fail(err)
	}

	view.Success()
	return nil
}

func channelLabel(sub *models.AppSlackChannelSubscription) string {
	if sub.ChannelName != "" {
		return sub.ChannelName
	}
	return sub.ChannelID
}

// describeSubscriptionMatch renders a match predicate the way the dashboard
// and Slack modal do ("Installs: not monitor=false"). Mirrors
// client/components/match/types.ts describeMatch and
// services/ctl-api/internal/app/slack/service/subscribe_modal.go describeMatch.
func describeSubscriptionMatch(match any) string {
	if match == nil {
		return "Org-wide"
	}
	b, err := json.Marshal(match)
	if err != nil {
		return "custom"
	}
	var m labels.SubscriptionMatch
	if err := json.Unmarshal(b, &m); err != nil {
		return "custom"
	}

	var parts []string
	for _, entry := range []struct {
		kind string
		tm   *labels.TargetMatch
	}{
		{"installs", m.Installs},
		{"components", m.Components},
		{"actions", m.Actions},
		{"app branches", m.AppBranches},
	} {
		if entry.tm == nil {
			continue
		}
		var kindParts []string
		if len(entry.tm.IDs) > 0 {
			kindParts = append(kindParts, fmt.Sprintf("%d ids", len(entry.tm.IDs)))
		}
		if entry.tm.Selector != nil {
			if inc := labelsToQueryString(entry.tm.Selector.MatchLabels); inc != "" {
				kindParts = append(kindParts, inc)
			}
			if exc := labelsToQueryString(entry.tm.Selector.NotMatchLabels); exc != "" {
				kindParts = append(kindParts, "not "+exc)
			}
		}
		if len(kindParts) == 0 {
			kindParts = append(kindParts, "any")
		}
		parts = append(parts, fmt.Sprintf("%s: %s", entry.kind, strings.Join(kindParts, "; ")))
	}
	if len(parts) == 0 {
		return "Org-wide"
	}
	return strings.Join(parts, "; ")
}

func labelsToQueryString(l labels.Labels) string {
	if len(l) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(l))
	for k, v := range l {
		if v == "*" {
			pairs = append(pairs, k)
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}
