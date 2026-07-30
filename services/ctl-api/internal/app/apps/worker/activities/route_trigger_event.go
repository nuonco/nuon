package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RouteTriggerEventRequest struct {
	EventID                string `validate:"required"`
	ReplayID               string
	RoutingGenerationToken string
}

type RouteTriggerEventResponse struct {
	Dispatches []TriggerEventDispatchRef `json:"dispatches"`
	Waiters    []EventRunbookWaiterRef   `json:"waiters"`
}

type EventRunbookWaiterRef struct {
	ID            string `json:"id"`
	OrgID         string `json:"org_id"`
	QueueSignalID string `json:"queue_signal_id"`
}

type TriggerEventDispatchRef struct {
	ID              string `json:"id"`
	AppID           string `json:"app_id"`
	GenerationToken string `json:"generation_token"`
}

// @temporal-gen-v2 activity
func (a *Activities) RouteTriggerEvent(ctx context.Context, req RouteTriggerEventRequest) (*RouteTriggerEventResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid route trigger event request: %w", err)
	}
	if req.ReplayID != "" {
		return a.routeTriggerEventReplay(ctx, req)
	}

	resp := &RouteTriggerEventResponse{}
	if req.RoutingGenerationToken == "" {
		routingGenerationToken := "legacy-" + req.EventID
		res := a.db.WithContext(ctx).Model(&app.TriggerEvent{}).
			Where(app.TriggerEvent{ID: req.EventID}).
			Where(clause.Eq{Column: "routing_generation_token", Value: nil}).
			Update("routing_generation_token", routingGenerationToken)
		if res.Error != nil {
			return nil, fmt.Errorf("unable to claim legacy trigger event: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			var claimed app.TriggerEvent
			if err := a.db.WithContext(ctx).Select("routing_generation_token").Where(app.TriggerEvent{ID: req.EventID}).First(&claimed).Error; err != nil {
				return nil, fmt.Errorf("unable to inspect legacy trigger event claim: %w", err)
			}
			if claimed.RoutingGenerationToken == nil || *claimed.RoutingGenerationToken != routingGenerationToken {
				return resp, nil
			}
		}
		req.RoutingGenerationToken = routingGenerationToken
	}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		var event app.TriggerEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Trigger").
			Where(app.TriggerEvent{ID: req.EventID}).
			First(&event).Error; err != nil {
			return fmt.Errorf("unable to get event: %w", err)
		}
		if event.RoutingGenerationToken == nil || *event.RoutingGenerationToken != req.RoutingGenerationToken {
			return nil
		}

		if event.RoutingStatus == app.EventRoutingStatusRejected {
			return nil
		}
		if event.RoutingStatus == app.EventRoutingStatusMatched || event.RoutingStatus == app.EventRoutingStatusIgnored {
			var dispatches []app.EventDispatch
			if err := tx.Where(app.EventDispatch{TriggerEventID: event.ID, OrgID: event.OrgID}).Find(&dispatches).Error; err != nil {
				return err
			}
			for i := range dispatches {
				if dispatches[i].ReplayID == nil && dispatches[i].Status != app.EventDispatchStatusDeadLettered && dispatches[i].Status != app.EventDispatchStatusCancelled {
					resp.Dispatches = append(resp.Dispatches, TriggerEventDispatchRef{ID: dispatches[i].ID, AppID: dispatches[i].AppID, GenerationToken: dispatches[i].GenerationToken})
				}
			}
			if err := appendPendingWaiters(tx, &event, resp); err != nil {
				return err
			}
			return nil
		}
		if event.Trigger.ID == "" {
			return markEventRejected(tx, &event, req.RoutingGenerationToken, "trigger was deleted")
		}
		var org app.Org
		if err := tx.Select("id", "features").Where(app.Org{ID: event.OrgID}).First(&org).Error; err != nil {
			return fmt.Errorf("unable to check triggers feature: %w", err)
		}
		if !org.Features[string(app.OrgFeatureTriggers)] {
			return markEventRejected(tx, &event, req.RoutingGenerationToken, "triggers feature is not enabled")
		}

		if err := eventGeneration(tx, &event, req.RoutingGenerationToken).Updates(map[string]any{
			"routing_status":       app.EventRoutingStatusRouting,
			"routing_error":        "",
			"routing_started_at":   now,
			"routing_completed_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("unable to mark event routing: %w", err)
		}

		var activeConfigs []app.AppConfig
		if err := tx.Where(app.AppConfig{OrgID: event.OrgID, Status: app.AppConfigStatusActive}).
			Order("created_at DESC").
			Find(&activeConfigs).Error; err != nil {
			return fmt.Errorf("unable to get active app configs: %w", err)
		}
		activeConfigIDs := app.ActiveTriggerConfigIDs(activeConfigs)
		activeConfigIDList := make([]string, 0, len(activeConfigIDs))
		for _, configID := range activeConfigIDs {
			activeConfigIDList = append(activeConfigIDList, configID)
		}

		var rules []app.TriggerRule
		if len(activeConfigIDList) != 0 {
			if err := tx.Where(app.TriggerRule{
				TriggerID: event.TriggerID,
				OrgID:     event.OrgID,
				Enabled:   true,
			}).Where(map[string]any{"app_config_id": activeConfigIDList}).Find(&rules).Error; err != nil {
				return fmt.Errorf("unable to list trigger rules: %w", err)
			}
		}

		payload, err := decodeTriggerEventPayload(event.Payload)
		if err != nil {
			return fmt.Errorf("unable to decode event payload: %w", err)
		}
		var trigger app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: event.TriggerID, OrgID: event.OrgID}).First(&trigger).Error; err != nil {
			return err
		}
		var waiters []app.EventRunbookWaiter
		if err := tx.Where(app.EventRunbookWaiter{OrgID: event.OrgID, TriggerID: event.TriggerID, Status: app.EventRunbookWaiterStatusActive}).Where(clause.Expr{SQL: "activated_at <= ?", Vars: []any{event.ReceivedAt}}).Find(&waiters).Error; err != nil {
			return err
		}
		for i := range waiters {
			w := &waiters[i]
			if !waiterMatchesEvent(w, &event, payload, http.Header(event.Headers)) {
				continue
			}
			matchedAt := time.Now().UTC()
			res := tx.Model(&app.EventRunbookWaiter{}).Where(app.EventRunbookWaiter{ID: w.ID, OrgID: event.OrgID, Status: app.EventRunbookWaiterStatusActive}).Updates(map[string]any{"status": app.EventRunbookWaiterStatusMatched, "matched_event_id": event.ID, "matched_at": matchedAt})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 0 {
				resp.Waiters = append(resp.Waiters, EventRunbookWaiterRef{ID: w.ID, OrgID: w.OrgID, QueueSignalID: w.QueueSignalID})
			}
		}

		matchCount := 0
		dispatchCount := 0
		explanations := make([]app.TriggerRuleEvaluation, 0, len(rules))
		for i := range rules {
			rule := &rules[i]
			if rule.SuspendedAt != nil {
				continue
			}
			matched, explanation := evaluateRule(rule, event.EventType, payload, http.Header(event.Headers))
			explanations = append(explanations, explanation)
			if !matched {
				continue
			}
			matchCount++
			persisted, err := createRuleDispatch(tx, &event, rule, payload, req.ReplayID)
			if err != nil {
				return err
			}
			dispatchCount++
			if persisted.Status != app.EventDispatchStatusDeadLettered {
				resp.Dispatches = append(resp.Dispatches, TriggerEventDispatchRef{ID: persisted.ID, AppID: persisted.AppID, GenerationToken: persisted.GenerationToken})
			}
		}

		return markEventRouted(tx, &event, req.RoutingGenerationToken, matchCount, len(resp.Waiters), dispatchCount, explanations)
	})
	if err != nil {
		completedAt := time.Now().UTC()
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		statusErr := a.db.WithContext(cleanupCtx).Model(&app.TriggerEvent{}).
			Where(app.TriggerEvent{ID: req.EventID, RoutingGenerationToken: &req.RoutingGenerationToken}).
			Not(app.TriggerEvent{RoutingStatus: app.EventRoutingStatusMatched}).
			Not(app.TriggerEvent{RoutingStatus: app.EventRoutingStatusIgnored}).
			Not(app.TriggerEvent{RoutingStatus: app.EventRoutingStatusRejected}).
			Updates(map[string]any{
				"routing_status":       app.EventRoutingStatusRoutingFailed,
				"routing_error":        err.Error(),
				"routing_completed_at": completedAt,
			}).Error
		if statusErr != nil {
			return nil, errors.Join(err, fmt.Errorf("unable to persist routing failure: %w", statusErr))
		}
		return nil, err
	}
	if err := a.ensureDispatchQueues(ctx, resp.Dispatches); err != nil {
		return nil, err
	}
	return resp, nil
}

// routeTriggerEventReplay evaluates the event against the currently active
// rules and creates replay-scoped dispatches. It never mutates the original
// event's routing record: the ledger row permanently describes initial
// routing, while replay outcomes are observable through dispatch rows
// carrying the replay ID.
func (a *Activities) routeTriggerEventReplay(ctx context.Context, req RouteTriggerEventRequest) (*RouteTriggerEventResponse, error) {
	resp := &RouteTriggerEventResponse{}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var event app.TriggerEvent
		if err := tx.Preload("Trigger").Where(app.TriggerEvent{ID: req.EventID}).First(&event).Error; err != nil {
			return fmt.Errorf("unable to get event: %w", err)
		}
		if event.RoutingStatus == app.EventRoutingStatusRejected {
			return nil
		}
		if event.Trigger.ID == "" {
			return errors.New("trigger was deleted")
		}
		var org app.Org
		if err := tx.Select("id", "features").Where(app.Org{ID: event.OrgID}).First(&org).Error; err != nil {
			return fmt.Errorf("unable to check triggers feature: %w", err)
		}
		if !org.Features[string(app.OrgFeatureTriggers)] {
			return errors.New("triggers feature is not enabled")
		}

		var activeConfigs []app.AppConfig
		if err := tx.Where(app.AppConfig{OrgID: event.OrgID, Status: app.AppConfigStatusActive}).
			Order("created_at DESC").
			Find(&activeConfigs).Error; err != nil {
			return fmt.Errorf("unable to get active app configs: %w", err)
		}
		activeConfigIDs := app.ActiveTriggerConfigIDs(activeConfigs)
		activeConfigIDList := make([]string, 0, len(activeConfigIDs))
		for _, configID := range activeConfigIDs {
			activeConfigIDList = append(activeConfigIDList, configID)
		}

		var rules []app.TriggerRule
		if len(activeConfigIDList) != 0 {
			if err := tx.Where(app.TriggerRule{
				TriggerID: event.TriggerID,
				OrgID:     event.OrgID,
				Enabled:   true,
			}).Where(map[string]any{"app_config_id": activeConfigIDList}).Find(&rules).Error; err != nil {
				return fmt.Errorf("unable to list trigger rules: %w", err)
			}
		}

		payload, err := decodeTriggerEventPayload(event.Payload)
		if err != nil {
			return fmt.Errorf("unable to decode event payload: %w", err)
		}
		for i := range rules {
			rule := &rules[i]
			if rule.SuspendedAt != nil {
				continue
			}
			matched, _ := evaluateRule(rule, event.EventType, payload, http.Header(event.Headers))
			if !matched {
				continue
			}
			persisted, err := createRuleDispatch(tx, &event, rule, payload, req.ReplayID)
			if err != nil {
				return err
			}
			if persisted.Status != app.EventDispatchStatusDeadLettered {
				resp.Dispatches = append(resp.Dispatches, TriggerEventDispatchRef{ID: persisted.ID, AppID: persisted.AppID, GenerationToken: persisted.GenerationToken})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := a.ensureDispatchQueues(ctx, resp.Dispatches); err != nil {
		return nil, err
	}
	return resp, nil
}

func createRuleDispatch(tx *gorm.DB, event *app.TriggerEvent, rule *app.TriggerRule, payload any, replayID string) (*app.EventDispatch, error) {
	dispatch := app.EventDispatch{
		CreatedByID:    event.Trigger.CreatedByID,
		OrgID:          event.OrgID,
		AppID:          rule.AppID,
		TriggerEventID: event.ID,
		TriggerRuleID:  rule.ID,
		ReplayID:       replayIDPtr(replayID),
		IdempotencyKey: triggerEventDispatchKey(event.ID, rule.ID, replayID),
		TargetType:     rule.TargetType,
		TargetID:       pointerValue(rule.AppBranchID),
		Status:         app.EventDispatchStatusPending,
	}
	if rule.TargetType == app.TriggerTargetTypeRunbook {
		if err := resolveRunbookDispatch(tx, rule, payload, &dispatch); err != nil {
			now := time.Now().UTC()
			dispatch.Status = app.EventDispatchStatusDeadLettered
			dispatch.Error = err.Error()
			dispatch.FailedAt = &now
		}
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "idempotency_key"}},
		DoNothing: true,
	}).Create(&dispatch).Error; err != nil {
		return nil, fmt.Errorf("unable to create event dispatch for rule %q: %w", rule.Name, err)
	}
	var persisted app.EventDispatch
	if err := tx.Where(app.EventDispatch{IdempotencyKey: dispatch.IdempotencyKey}).First(&persisted).Error; err != nil {
		return nil, fmt.Errorf("unable to get event dispatch for rule %q: %w", rule.Name, err)
	}
	return &persisted, nil
}

// ensureDispatchQueues guarantees the per-app trigger queues exist before the
// trigger-event signal fans dispatches out to them.
func (a *Activities) ensureDispatchQueues(ctx context.Context, dispatches []TriggerEventDispatchRef) error {
	ensured := make(map[string]struct{}, len(dispatches))
	for _, dispatch := range dispatches {
		if _, ok := ensured[dispatch.AppID]; ok {
			continue
		}
		if _, err := a.helpers.EnsureAppTriggerQueue(ctx, dispatch.AppID); err != nil {
			return err
		}
		ensured[dispatch.AppID] = struct{}{}
	}
	return nil
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func resolveRunbookDispatch(db *gorm.DB, rule *app.TriggerRule, payload any, dispatch *app.EventDispatch) error {
	if rule.RunbookID == nil {
		return errors.New("runbook target has no runbook")
	}
	var install app.Install
	if err := db.Where(app.Install{OrgID: rule.OrgID, AppID: rule.AppID, Name: rule.InstallName}).First(&install).Error; err != nil {
		return fmt.Errorf("resolve install %q: %w", rule.InstallName, err)
	}
	var installRunbook app.InstallRunbook
	if err := db.Where(app.InstallRunbook{OrgID: rule.OrgID, InstallID: install.ID, RunbookID: *rule.RunbookID}).First(&installRunbook).Error; err != nil {
		return fmt.Errorf("resolve install runbook: %w", err)
	}
	var runbookConfig app.RunbookConfig
	if err := db.Preload("Inputs").Where(app.RunbookConfig{OrgID: rule.OrgID, RunbookID: *rule.RunbookID, AppConfigID: install.AppConfigID}).First(&runbookConfig).Error; err != nil {
		return fmt.Errorf("resolve pinned runbook config: %w", err)
	}
	declared := make(map[string]app.RunbookInput, len(runbookConfig.Inputs))
	for _, input := range runbookConfig.Inputs {
		declared[input.Name] = input
	}
	for name := range rule.InputMappings {
		if _, ok := declared[name]; !ok {
			return fmt.Errorf("mapped input %q is not declared by pinned runbook config", name)
		}
	}
	for _, input := range runbookConfig.Inputs {
		if input.Required && input.Default == "" {
			if _, ok := rule.InputMappings[input.Name]; !ok {
				return fmt.Errorf("required input %q is not mapped", input.Name)
			}
		}
	}
	mapped, err := mapTriggerInputs(rule.InputMappings, payload)
	if err != nil {
		return err
	}
	for _, input := range runbookConfig.Inputs {
		if input.Required && input.Default == "" && mapped[input.Name] == "" {
			return fmt.Errorf("required input %q mapped to an empty value", input.Name)
		}
	}
	dispatch.TargetID = installRunbook.ID
	dispatch.RunbookConfigID = &runbookConfig.ID
	dispatch.MappedInputs = mapped
	return nil
}

func mapTriggerInputs(mappings map[string]string, payload any) (map[string]string, error) {
	mapped := make(map[string]string, len(mappings))
	for name, rawPath := range mappings {
		path, err := eventfilter.ParsePath(rawPath, false)
		if err != nil {
			return nil, fmt.Errorf("parse mapping for %q: %w", name, err)
		}
		selected := path.Select(payload)
		if len(selected) != 1 {
			return nil, fmt.Errorf("mapping for %q selected %d values; expected exactly one", name, len(selected))
		}
		if selected[0] == nil {
			return nil, fmt.Errorf("mapping for %q selected null", name)
		}
		switch value := selected[0].(type) {
		case string:
			mapped[name] = value
		case json.Number:
			mapped[name] = value.String()
		case bool:
			mapped[name] = fmt.Sprint(value)
		default:
			return nil, fmt.Errorf("mapping for %q selected a non-scalar value", name)
		}
	}
	return mapped, nil
}

func replayIDPtr(replayID string) *string {
	if replayID == "" {
		return nil
	}
	return &replayID
}

func triggerEventDispatchKey(eventID, ruleID, replayID string) string {
	key := eventID + ":" + ruleID
	if replayID != "" {
		key += ":" + replayID
	}
	return key
}

func eventGeneration(tx *gorm.DB, event *app.TriggerEvent, routingGenerationToken string) *gorm.DB {
	return tx.Model(event).Where(app.TriggerEvent{RoutingGenerationToken: &routingGenerationToken})
}

func markEventRejected(tx *gorm.DB, event *app.TriggerEvent, routingGenerationToken, reason string) error {
	completedAt := time.Now().UTC()
	return eventGeneration(tx, event, routingGenerationToken).Updates(map[string]any{
		"routing_status":       app.EventRoutingStatusRejected,
		"routing_error":        reason,
		"routing_completed_at": completedAt,
		"match_count":          0,
		"waiter_match_count":   0,
		"dispatch_count":       0,
	}).Error
}

func markEventRouted(tx *gorm.DB, event *app.TriggerEvent, routingGenerationToken string, matchCount, waiterMatchCount, dispatchCount int, explanations []app.TriggerRuleEvaluation) error {
	completedAt := time.Now().UTC()
	status := routedEventStatus(matchCount, waiterMatchCount)
	explanations, explanationsTruncated := boundedExplanations(explanations)
	encodedExplanations, err := json.Marshal(explanations)
	if err != nil {
		return fmt.Errorf("encode event match explanations: %w", err)
	}
	return eventGeneration(tx, event, routingGenerationToken).Updates(map[string]any{
		"routing_status":         status,
		"routing_error":          "",
		"routing_completed_at":   completedAt,
		"match_count":            matchCount,
		"waiter_match_count":     waiterMatchCount,
		"dispatch_count":         dispatchCount,
		"match_explanations":     string(encodedExplanations),
		"explanations_truncated": explanationsTruncated,
	}).Error
}

func routedEventStatus(ruleMatchCount, waiterMatchCount int) app.EventRoutingStatus {
	if ruleMatchCount == 0 && waiterMatchCount == 0 {
		return app.EventRoutingStatusIgnored
	}
	return app.EventRoutingStatusMatched
}

func boundedExplanations(explanations []app.TriggerRuleEvaluation) ([]app.TriggerRuleEvaluation, bool) {
	const maxBytes = 256 * 1024
	bounded := make([]app.TriggerRuleEvaluation, 0, len(explanations))
	for _, explanation := range explanations {
		candidate := append(bounded, explanation)
		encoded, err := json.Marshal(candidate)
		if err != nil || len(encoded) > maxBytes {
			return bounded, true
		}
		bounded = candidate
	}
	return bounded, false
}

func decodeTriggerEventPayload(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func ruleMatchesEvent(rule *app.TriggerRule, eventType string, payload any, headers ...http.Header) bool {
	var requestHeaders http.Header
	if len(headers) != 0 {
		requestHeaders = headers[0]
	}
	matched, _ := evaluateRule(rule, eventType, payload, requestHeaders)
	return matched
}

func evaluateRule(rule *app.TriggerRule, eventType string, payload any, headers http.Header) (bool, app.TriggerRuleEvaluation) {
	explanation := app.TriggerRuleEvaluation{
		RuleID: rule.ID, RuleName: rule.Name, AppID: rule.AppID, EventType: eventType,
		AllowedEventTypes: []string(rule.EventTypes), EventTypeMatched: true,
	}
	if len(rule.EventTypes) != 0 {
		explanation.EventTypeMatched = false
		for _, allowed := range rule.EventTypes {
			if eventType == allowed {
				explanation.EventTypeMatched = true
				break
			}
		}
	}
	filtersMatched := true
	for _, filter := range rule.Filters {
		filterExplanation := app.TriggerFilterEvaluation{
			From: filter.From, Path: filter.Path, Op: filter.Op, Expected: filter.Value,
		}
		compiled, err := eventfilter.Compile(eventfilter.Filter{
			From: eventfilter.Source(filter.From), Path: filter.Path,
			Op: eventfilter.Operator(filter.Op), Value: filter.Value,
		})
		if err != nil {
			filterExplanation.Error = err.Error()
			filtersMatched = false
		} else {
			result := compiled.Evaluate(payload, headers)
			filterExplanation.Matched = result.Matched
			filterExplanation.Selected, filterExplanation.Truncated = explanationValues(result.Selected)
			filtersMatched = filtersMatched && result.Matched
		}
		explanation.Filters = append(explanation.Filters, filterExplanation)
	}
	explanation.Matched = explanation.EventTypeMatched && filtersMatched
	return explanation.Matched, explanation
}

func eventSetMatches(eventTypes []string, filters []app.TriggerFilter, eventType string, payload any, headers http.Header) bool {
	rule := &app.TriggerRule{EventTypes: eventTypes, Filters: filters}
	matched, _ := evaluateRule(rule, eventType, payload, headers)
	return matched
}

func waiterMatchesEvent(waiter *app.EventRunbookWaiter, event *app.TriggerEvent, payload any, headers http.Header) bool {
	if waiter.ActivatedAt.After(event.ReceivedAt) {
		return false
	}
	return eventSetMatches([]string(waiter.EventTypes), waiter.Filters, event.EventType, payload, headers)
}

func appendPendingWaiters(tx *gorm.DB, event *app.TriggerEvent, resp *RouteTriggerEventResponse) error {
	var waiters []app.EventRunbookWaiter
	if err := tx.Where(app.EventRunbookWaiter{MatchedEventID: &event.ID, OrgID: event.OrgID, Status: app.EventRunbookWaiterStatusMatched}).Where(clause.Eq{Column: "notified_at", Value: nil}).Find(&waiters).Error; err != nil {
		return err
	}
	for _, waiter := range waiters {
		resp.Waiters = append(resp.Waiters, EventRunbookWaiterRef{ID: waiter.ID, OrgID: waiter.OrgID, QueueSignalID: waiter.QueueSignalID})
	}
	return nil
}

func explanationValues(values []any) ([]any, bool) {
	const maxValues = 10
	const maxValueSize = 2048
	truncated := len(values) > maxValues
	if truncated {
		values = values[:maxValues]
	}
	explained := make([]any, len(values))
	for i, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil || len(encoded) > maxValueSize {
			explained[i] = "<value omitted>"
			truncated = true
			continue
		}
		explained[i] = value
	}
	return explained, truncated
}
