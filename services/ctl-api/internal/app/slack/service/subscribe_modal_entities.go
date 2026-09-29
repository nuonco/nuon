package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// why: The action_id constants for the subscribe modal's per-resource-kind
// multi_external_select pickers (subscribeEntitiesActionIDInstalls /
// subscribeEntitiesActionIDComponents / subscribeEntitiesActionIDActions)
// are declared in subscribe_modal.go alongside the modal's other Block-Kit
// identifiers; this file only references them so the dispatcher and the
// modal cannot drift.

func (s *service) resolveSubscribeSuggestionOrgLink(
	ctx context.Context,
	payload slackInteractionPayload,
	logFieldPrefix string,
) (app.SlackOrgLink, bool) {
	state, err := decodeSubscribeModalState(payload.View.PrivateMetadata)
	if err != nil {
		s.l.Warn(logFieldPrefix+": decode state failed", zap.Error(err))
		return app.SlackOrgLink{}, false
	}
	if state.TeamID != payload.Team.ID {
		s.l.Warn(logFieldPrefix+": team_id mismatch",
			zap.String("state_team_id", state.TeamID),
			zap.String("payload_team_id", payload.Team.ID))
		return app.SlackOrgLink{}, false
	}

	render := readSubscribeRenderStateFromPayload(payload)
	if render.OrgLinkID == "" {
		return app.SlackOrgLink{}, false
	}

	var link app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:     render.OrgLinkID,
			TeamID: state.TeamID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		First(&link).Error; err != nil {
		s.l.Warn(logFieldPrefix+": org link not trusted", zap.Error(err))
		return app.SlackOrgLink{}, false
	}
	return link, true
}

const suggestionOptionsMaxResults = 100

func buildSuggestionOptions(items []suggestionItem) map[string]any {
	options := make([]any, 0, len(items))
	for _, it := range items {
		options = append(options, map[string]any{
			"text":  map[string]any{"type": "plain_text", "text": truncatePlainText(it.Name, 75)},
			"value": it.ID,
		})
	}
	return map[string]any{"options": options}
}

type suggestionItem struct {
	ID   string
	Name string
}

// why: handleSubscribeModalAppsBlockSuggestion serves the app picker's
// external_select dynamic options. The app picker is only rendered when
// the user has chosen a kind that scopes to an app (components /
// actions); selecting an app re-renders the entity picker against that
// app's children. Like the entity handlers, the org id used to filter
// apps comes from the workspace-trusted SlackOrgLink, never from
// user-supplied input.
func (s *service) handleSubscribeModalAppsBlockSuggestion(
	ctx context.Context,
	payload slackInteractionPayload,
) map[string]any {
	empty := map[string]any{"options": []any{}}

	link, ok := s.resolveSubscribeSuggestionOrgLink(ctx, payload, "subscribe modal suggest apps")
	if !ok {
		return empty
	}

	tx := s.db.WithContext(ctx).
		Model(&app.App{}).
		Where(&app.App{OrgID: link.OrgID}).
		Order("name ASC").
		Limit(suggestionOptionsMaxResults)
	if q := payload.Value; q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	var apps []app.App
	if err := tx.Find(&apps).Error; err != nil {
		s.l.Warn("subscribe modal suggest apps: list failed", zap.Error(err))
		return empty
	}

	items := make([]suggestionItem, 0, len(apps))
	for _, a := range apps {
		items = append(items, suggestionItem{ID: a.ID, Name: a.Name})
	}
	return buildSuggestionOptions(items)
}

func (s *service) resolveSubscribeSuggestionApp(
	ctx context.Context,
	payload slackInteractionPayload,
	logFieldPrefix string,
) (app.SlackOrgLink, app.App, bool) {
	link, ok := s.resolveSubscribeSuggestionOrgLink(ctx, payload, logFieldPrefix)
	if !ok {
		return app.SlackOrgLink{}, app.App{}, false
	}
	render := readSubscribeRenderStateFromPayload(payload)
	if render.AppID == "" {
		return app.SlackOrgLink{}, app.App{}, false
	}
	var a app.App
	if err := s.db.WithContext(ctx).
		Where(&app.App{ID: render.AppID, OrgID: link.OrgID}).
		First(&a).Error; err != nil {
		s.l.Warn(logFieldPrefix+": app not in trusted org", zap.Error(err))
		return app.SlackOrgLink{}, app.App{}, false
	}
	return link, a, true
}

func (s *service) handleSubscribeModalComponentsBlockSuggestion(
	ctx context.Context,
	payload slackInteractionPayload,
) map[string]any {
	empty := map[string]any{"options": []any{}}

	_, a, ok := s.resolveSubscribeSuggestionApp(ctx, payload, "subscribe modal suggest components")
	if !ok {
		return empty
	}

	tx := s.db.WithContext(ctx).
		Model(&app.Component{}).
		Where(&app.Component{AppID: a.ID, OrgID: a.OrgID}).
		Order("name ASC").
		Limit(suggestionOptionsMaxResults)
	if q := payload.Value; q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	var components []app.Component
	if err := tx.Find(&components).Error; err != nil {
		s.l.Warn("subscribe modal suggest components: list failed", zap.Error(err))
		return empty
	}

	items := make([]suggestionItem, 0, len(components))
	for _, c := range components {
		items = append(items, suggestionItem{ID: c.ID, Name: c.Name})
	}
	return buildSuggestionOptions(items)
}

func (s *service) handleSubscribeModalAppBranchesBlockSuggestion(
	ctx context.Context,
	payload slackInteractionPayload,
) map[string]any {
	empty := map[string]any{"options": []any{}}

	_, a, ok := s.resolveSubscribeSuggestionApp(ctx, payload, "subscribe modal suggest app branches")
	if !ok {
		return empty
	}

	tx := s.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Where(&app.AppBranch{AppID: a.ID}).
		Order("name ASC").
		Limit(suggestionOptionsMaxResults)
	if q := payload.Value; q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	var branches []app.AppBranch
	if err := tx.Find(&branches).Error; err != nil {
		s.l.Warn("subscribe modal suggest app branches: list failed", zap.Error(err))
		return empty
	}

	items := make([]suggestionItem, 0, len(branches))
	for _, b := range branches {
		items = append(items, suggestionItem{ID: b.ID, Name: b.Name})
	}
	return buildSuggestionOptions(items)
}

func (s *service) handleSubscribeModalActionsBlockSuggestion(
	ctx context.Context,
	payload slackInteractionPayload,
) map[string]any {
	empty := map[string]any{"options": []any{}}

	_, a, ok := s.resolveSubscribeSuggestionApp(ctx, payload, "subscribe modal suggest actions")
	if !ok {
		return empty
	}

	tx := s.db.WithContext(ctx).
		Model(&app.ActionWorkflow{}).
		Where(&app.ActionWorkflow{AppID: a.ID}).
		Order("name ASC").
		Limit(suggestionOptionsMaxResults)
	if q := payload.Value; q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	var workflows []app.ActionWorkflow
	if err := tx.Find(&workflows).Error; err != nil {
		s.l.Warn("subscribe modal suggest actions: list failed", zap.Error(err))
		return empty
	}

	items := make([]suggestionItem, 0, len(workflows))
	for _, w := range workflows {
		items = append(items, suggestionItem{ID: w.ID, Name: w.Name})
	}
	return buildSuggestionOptions(items)
}
