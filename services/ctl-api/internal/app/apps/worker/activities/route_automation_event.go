package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type RouteAutomationEventRequest struct {
	EventID string `validate:"required"`
}

type RouteAutomationEventResponse struct {
	DispatchIDs []string `json:"dispatch_ids"`
}

// @temporal-gen-v2 activity
func (a *Activities) RouteAutomationEvent(ctx context.Context, req RouteAutomationEventRequest) (*RouteAutomationEventResponse, error) {
	if err := a.v.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid route automation event request: %w", err)
	}

	resp := &RouteAutomationEventResponse{}
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		var event app.EventSourceEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("EventSource").
			Where(app.EventSourceEvent{ID: req.EventID}).
			First(&event).Error; err != nil {
			return fmt.Errorf("unable to get event: %w", err)
		}

		if event.RoutingStatus == app.EventRoutingStatusRouted {
			return tx.Model(&app.EventDispatch{}).
				Where(app.EventDispatch{EventSourceEventID: event.ID}).
				Pluck("id", &resp.DispatchIDs).Error
		}

		if err := tx.Model(&event).Updates(map[string]any{
			"routing_status":       app.EventRoutingStatusRouting,
			"routing_error":        "",
			"routing_started_at":   now,
			"routing_completed_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("unable to mark event routing: %w", err)
		}

		var activeConfigs []app.AppConfig
		if err := tx.Where(app.AppConfig{AppID: event.AppID, Status: app.AppConfigStatusActive}).
			Order("created_at DESC").
			Find(&activeConfigs).Error; err != nil {
			return fmt.Errorf("unable to get active app configs: %w", err)
		}
		var activeConfig *app.AppConfig
		for i := range activeConfigs {
			if activeConfigs[i].Labels["source"] != string(app.AppBranchRunTypeGitPreview) {
				activeConfig = &activeConfigs[i]
				break
			}
		}
		if activeConfig == nil {
			return markEventRouted(tx, &event, 0, 0)
		}

		var rules []app.EventAutomationRule
		if err := tx.Where(app.EventAutomationRule{
			AppConfigID:   activeConfig.ID,
			EventSourceID: event.EventSourceID,
			Enabled:       true,
		}).Find(&rules).Error; err != nil {
			return fmt.Errorf("unable to list event automation rules: %w", err)
		}

		payload, err := decodeAutomationPayload(event.Payload)
		if err != nil {
			return fmt.Errorf("unable to decode event payload: %w", err)
		}

		matchCount := 0
		for i := range rules {
			rule := &rules[i]
			if rule.SuspendedAt != nil || !ruleMatchesEvent(rule, event.EventType, payload) {
				continue
			}
			matchCount++
			dispatch := app.EventDispatch{
				CreatedByID:           event.EventSource.CreatedByID,
				OrgID:                 event.OrgID,
				AppID:                 event.AppID,
				EventSourceEventID:    event.ID,
				EventAutomationRuleID: rule.ID,
				IdempotencyKey:        event.ID + ":" + rule.ID,
				TargetType:            rule.TargetType,
				TargetID:              rule.AppBranchID,
				Status:                app.EventDispatchStatusPending,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "idempotency_key"}},
				DoNothing: true,
			}).Create(&dispatch).Error; err != nil {
				return fmt.Errorf("unable to create event dispatch for rule %q: %w", rule.Name, err)
			}
			var persisted app.EventDispatch
			if err := tx.Where(app.EventDispatch{IdempotencyKey: dispatch.IdempotencyKey}).First(&persisted).Error; err != nil {
				return fmt.Errorf("unable to get event dispatch for rule %q: %w", rule.Name, err)
			}
			resp.DispatchIDs = append(resp.DispatchIDs, persisted.ID)
		}

		return markEventRouted(tx, &event, matchCount, len(resp.DispatchIDs))
	})
	if err != nil {
		completedAt := time.Now().UTC()
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		statusErr := a.db.WithContext(cleanupCtx).Model(&app.EventSourceEvent{}).
			Where(app.EventSourceEvent{ID: req.EventID}).
			Not(app.EventSourceEvent{RoutingStatus: app.EventRoutingStatusRouted}).
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
	return resp, nil
}

func markEventRouted(tx *gorm.DB, event *app.EventSourceEvent, matchCount, dispatchCount int) error {
	completedAt := time.Now().UTC()
	return tx.Model(event).Updates(map[string]any{
		"routing_status":       app.EventRoutingStatusRouted,
		"routing_error":        "",
		"routing_completed_at": completedAt,
		"match_count":          matchCount,
		"dispatch_count":       dispatchCount,
	}).Error
}

func decodeAutomationPayload(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func ruleMatchesEvent(rule *app.EventAutomationRule, eventType string, payload any) bool {
	eventTypeMatches := false
	for _, allowed := range rule.EventTypes {
		if eventType == allowed {
			eventTypeMatches = true
			break
		}
	}
	if !eventTypeMatches {
		return false
	}
	for _, filter := range rule.Filters {
		if filter.Op != app.EventAutomationFilterTypeEq {
			return false
		}
		actual, ok := resolveJSONPointer(payload, filter.Path)
		if !ok || !jsonValuesEqual(actual, filter.Value) {
			return false
		}
	}
	return true
}

func resolveJSONPointer(value any, pointer string) (any, bool) {
	current := value
	for _, encoded := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		key := strings.ReplaceAll(strings.ReplaceAll(encoded, "~1", "/"), "~0", "~")
		switch container := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = container[key]
			if !ok {
				return nil, false
			}
		case []any:
			if key != "0" && (key == "" || key[0] == '0') {
				return nil, false
			}
			for _, digit := range key {
				if digit < '0' || digit > '9' {
					return nil, false
				}
			}
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(container) {
				return nil, false
			}
			current = container[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func jsonValuesEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return reflect.DeepEqual(left, right)
	}
	var normalizedLeft, normalizedRight any
	leftDecoder := json.NewDecoder(bytes.NewReader(leftJSON))
	leftDecoder.UseNumber()
	rightDecoder := json.NewDecoder(bytes.NewReader(rightJSON))
	rightDecoder.UseNumber()
	if leftDecoder.Decode(&normalizedLeft) != nil || rightDecoder.Decode(&normalizedRight) != nil {
		return reflect.DeepEqual(left, right)
	}
	leftNumber, leftIsNumber := normalizedLeft.(json.Number)
	rightNumber, rightIsNumber := normalizedRight.(json.Number)
	if leftIsNumber && rightIsNumber {
		leftRat, leftOK := new(big.Rat).SetString(leftNumber.String())
		rightRat, rightOK := new(big.Rat).SetString(rightNumber.String())
		return leftOK && rightOK && leftRat.Cmp(rightRat) == 0
	}
	return reflect.DeepEqual(normalizedLeft, normalizedRight)
}
