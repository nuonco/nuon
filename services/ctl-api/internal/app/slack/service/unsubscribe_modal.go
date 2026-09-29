package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

const (
	unsubscribeModalCallbackID    = "nuon_unsubscribe_modal"
	unsubscribeRemoveActionID     = "nuon_unsubscribe_remove"
	unsubscribeRemoveBlockPrefix  = "nuon_unsubscribe_row_"
	unsubscribeEmptySectionBlock  = "nuon_unsubscribe_empty"
	unsubscribeHeaderSectionBlock = "nuon_unsubscribe_header"
)

type unsubscribeModalState struct {
	TeamID      string `json:"team_id"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

func buildUnsubscribeModalView(
	state unsubscribeModalState,
	subs []app.SlackChannelSubscription,
	links []app.SlackOrgLink,
) (map[string]any, error) {
	pm, err := encodeUnsubscribeModalState(state)
	if err != nil {
		return nil, err
	}

	channelLabel := state.ChannelID
	if state.ChannelName != "" {
		channelLabel = "#" + state.ChannelName
	}

	blocks := []any{
		map[string]any{
			"type":     "section",
			"block_id": unsubscribeHeaderSectionBlock,
			"text": map[string]any{
				"type": "mrkdwn",
				"text": fmt.Sprintf("Subscriptions for <#%s> (%s):", state.ChannelID, channelLabel),
			},
		},
	}

	if len(subs) == 0 {
		blocks = append(blocks, map[string]any{
			"type":     "section",
			"block_id": unsubscribeEmptySectionBlock,
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "_No active subscriptions for this channel._",
			},
		})
	} else {
		blocks = append(blocks, map[string]any{"type": "divider"})

		orgNameByLinkID := make(map[string]string, len(links))
		for _, l := range links {
			if l.Org.Name != "" {
				orgNameByLinkID[l.ID] = l.Org.Name
			} else {
				orgNameByLinkID[l.ID] = l.OrgID
			}
		}

		ordered := append([]app.SlackChannelSubscription(nil), subs...)
		sort.SliceStable(ordered, func(i, j int) bool {
			ni := orgNameByLinkID[ordered[i].OrgLinkID]
			nj := orgNameByLinkID[ordered[j].OrgLinkID]
			if ni != nj {
				return ni < nj
			}
			return ordered[i].ID < ordered[j].ID
		})

		for _, sub := range ordered {
			orgName, ok := orgNameByLinkID[sub.OrgLinkID]
			if !ok {
				orgName = sub.OrgID
			}
			blocks = append(blocks, map[string]any{
				"type":     "section",
				"block_id": unsubscribeRemoveBlockPrefix + sub.ID,
				"text": map[string]any{
					"type": "mrkdwn",
					"text": describeUnsubscribeRow(sub, orgName),
				},
				"accessory": map[string]any{
					"type":      "button",
					"action_id": unsubscribeRemoveActionID,
					"text":      map[string]any{"type": "plain_text", "text": "Remove"},
					"style":     "danger",
					"value":     sub.ID,
				},
			})
		}
	}

	return map[string]any{
		"type":             "modal",
		"callback_id":      unsubscribeModalCallbackID,
		"private_metadata": pm,
		"title":            map[string]any{"type": "plain_text", "text": "Unsubscribe from Nuon"},
		"close":            map[string]any{"type": "plain_text", "text": "Close"},
		"blocks":           blocks,
	}, nil
}

func describeUnsubscribeRow(sub app.SlackChannelSubscription, orgName string) string {
	scopeTag := describeMatch(sub.Match)
	filter := "specific events"
	if sub.Interests.AllEvents {
		filter = "all events"
	}
	return fmt.Sprintf("*%s* · %s · %s", orgName, scopeTag, filter)
}

func encodeUnsubscribeModalState(s unsubscribeModalState) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("encode unsubscribe modal state: %w", err)
	}
	return string(b), nil
}

func decodeUnsubscribeModalState(raw string) (unsubscribeModalState, error) {
	var s unsubscribeModalState
	if raw == "" {
		return s, errors.New("unsubscribe modal: empty private_metadata")
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return s, fmt.Errorf("decode unsubscribe modal state: %w", err)
	}
	return s, nil
}

func (s *service) openUnsubscribeModalForSlash(
	ctx context.Context,
	triggerID string,
	teamID, channelID, channelName string,
) error {
	install, err := s.lookupActiveInstallForTeam(ctx, teamID)
	if err != nil {
		return err
	}

	subs, links, err := s.loadChannelSubsForUnsubscribeModal(ctx, teamID, channelID)
	if err != nil {
		return err
	}

	state := unsubscribeModalState{
		TeamID:      teamID,
		ChannelID:   channelID,
		ChannelName: channelName,
	}
	view, err := buildUnsubscribeModalView(state, subs, links)
	if err != nil {
		return err
	}

	if _, err := s.slackClient.ViewsOpen(ctx, install.BotAccessToken, slackclient.ViewsOpenRequest{
		TriggerID: triggerID,
		View:      view,
	}); err != nil {
		return fmt.Errorf("unsubscribe modal: views.open: %w", err)
	}
	return nil
}

func (s *service) loadChannelSubsForUnsubscribeModal(
	ctx context.Context,
	teamID, channelID string,
) ([]app.SlackChannelSubscription, []app.SlackOrgLink, error) {
	var subs []app.SlackChannelSubscription
	if err := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Find(&subs).Error; err != nil {
		return nil, nil, fmt.Errorf("unsubscribe modal: list subs: %w", err)
	}

	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Preload("Org").
		Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		return nil, nil, fmt.Errorf("unsubscribe modal: list org links: %w", err)
	}
	return subs, links, nil
}

func (s *service) handleUnsubscribeModalBlockActions(ctx context.Context, payload slackInteractionPayload) {
	state, err := decodeUnsubscribeModalState(payload.View.PrivateMetadata)
	if err != nil {
		s.l.Warn("unsubscribe modal: decode state failed", zap.Error(err))
		return
	}
	if state.TeamID != payload.Team.ID {
		s.l.Warn("unsubscribe modal: team_id mismatch",
			zap.String("state_team_id", state.TeamID),
			zap.String("payload_team_id", payload.Team.ID))
		return
	}

	for _, a := range payload.Actions {
		if a.ActionID != unsubscribeRemoveActionID || a.Value == "" {
			continue
		}
		if err := s.softDeleteSubscriptionByIDForTeam(ctx, a.Value, state.TeamID); err != nil {
			s.l.Error("unsubscribe modal: delete sub failed",
				zap.Error(err),
				zap.String("sub_id", a.Value),
				zap.String("team_id", state.TeamID))
		}
	}

	if err := s.rerenderUnsubscribeModal(ctx, payload, state); err != nil {
		s.l.Warn("unsubscribe modal: re-render failed", zap.Error(err))
	}
}

// why: softDeleteSubscriptionByIDForTeam soft-deletes a single subscription,
// scoping the WHERE on TeamID so a forged sub id from another workspace
// can't reach across.
func (s *service) softDeleteSubscriptionByIDForTeam(ctx context.Context, subID, teamID string) error {
	res := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{ID: subID, TeamID: teamID}).
		Delete(&app.SlackChannelSubscription{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil
	}
	return nil
}

func (s *service) rerenderUnsubscribeModal(
	ctx context.Context,
	payload slackInteractionPayload,
	state unsubscribeModalState,
) error {
	install, err := s.lookupActiveInstallForTeam(ctx, state.TeamID)
	if err != nil {
		return err
	}
	subs, links, err := s.loadChannelSubsForUnsubscribeModal(ctx, state.TeamID, state.ChannelID)
	if err != nil {
		return err
	}
	view, err := buildUnsubscribeModalView(state, subs, links)
	if err != nil {
		return err
	}
	if _, err := s.slackClient.ViewsUpdate(ctx, install.BotAccessToken, slackclient.ViewsUpdateRequest{
		ViewID: payload.View.ID,
		Hash:   payload.View.Hash,
		View:   view,
	}); err != nil {
		return fmt.Errorf("unsubscribe modal: views.update: %w", err)
	}
	return nil
}
