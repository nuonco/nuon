package service

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type eventTypeFacet struct {
	EventType string `json:"event_type" gorm:"column:event_type"`
	Count     int64  `json:"count"`
}

type triggerRuleResponse struct {
	app.TriggerRule
	AppName       string `json:"app_name"`
	AppBranchName string `json:"app_branch_name,omitempty"`
	RunbookName   string `json:"runbook_name,omitempty"`
}

func parseEventListOrder(value string) (string, error) {
	order := strings.ToLower(strings.TrimSpace(value))
	if order == "" {
		return "desc", nil
	}
	if order != "asc" && order != "desc" {
		return "", errors.New("order must be asc or desc")
	}
	return order, nil
}

func (s *service) readTrigger(ctx *gin.Context) (*app.Trigger, bool) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	var trigger app.Trigger
	if err := s.db.WithContext(ctx).Select("id", "org_id").Where(app.Trigger{ID: ctx.Param("trigger_id"), OrgID: org.ID}).First(&trigger).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
		} else {
			ctx.Error(err)
		}
		return nil, false
	}
	return &trigger, true
}

// @ID ListTriggerEvents
// @Summary List events observed from an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Param event_type query string false "Exact event type"
// @Param outcome query string false "Event outcome: ok, ignored, rejected, processing, or failed"
// @Param query query string false "Search Nuon event ID or external provider ID"
// @Param received_after query string false "Only events received at or after this RFC3339 timestamp"
// @Param received_before query string false "Only events received at or before this RFC3339 timestamp"
// @Param order query string false "Sort direction: asc or desc (default desc)"
// @Param cursor query string false "Opaque pagination cursor"
// @Param limit query int false "Maximum events to return (max 100)"
// @Success 200 {object} eventListResponse
// @Failure 400 {object} stderr.ErrResponse
// @Router /v1/triggers/{trigger_id}/events [get]
func (s *service) ListTriggerEvents(ctx *gin.Context) {
	trigger, ok := s.readTrigger(ctx)
	if !ok {
		return
	}
	filters, err := parseEventListFilters(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := parseEventListOrder(ctx.Query("order"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cursor *eventListCursor
	if value := ctx.Query("cursor"); value != "" {
		cursor, err = decodeEventListCursor(value)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
	}
	s.listEvents(ctx, trigger.OrgID, trigger.ID, filters, cursor, order)
}

// @ID ListTriggerEventTypes
// @Summary List observed event types for an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {array} eventTypeFacet
// @Router /v1/triggers/{trigger_id}/event-types [get]
func (s *service) ListTriggerEventTypes(ctx *gin.Context) {
	trigger, ok := s.readTrigger(ctx)
	if !ok {
		return
	}
	var facets []eventTypeFacet
	if err := s.db.WithContext(ctx).Model(&app.TriggerEvent{}).
		Select("event_type", "COUNT(*) AS count").
		Where(app.TriggerEvent{OrgID: trigger.OrgID, TriggerID: trigger.ID}).
		Group("event_type").Order("count DESC, event_type ASC").Find(&facets).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, facets)
}

func activeTriggerRulesQuery(db *gorm.DB, trigger *app.Trigger, configIDs []string) *gorm.DB {
	return db.Where(app.TriggerRule{OrgID: trigger.OrgID, TriggerID: trigger.ID, Enabled: true}).
		Where(map[string]any{"app_config_id": configIDs}).
		Where(clause.Eq{Column: "suspended_at", Value: nil}).
		Where(clause.Eq{Column: "valid_to", Value: nil})
}

func (s *service) activeTriggerConfigIDs(ctx *gin.Context, orgID string) ([]string, error) {
	var configs []app.AppConfig
	if err := s.db.WithContext(ctx).Where(app.AppConfig{OrgID: orgID, Status: app.AppConfigStatusActive}).Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, err
	}
	active := app.ActiveTriggerConfigIDs(configs)
	ids := make([]string, 0, len(active))
	for _, id := range active {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *service) triggerRuleResponse(ctx *gin.Context, rule app.TriggerRule) (triggerRuleResponse, error) {
	response := triggerRuleResponse{TriggerRule: rule, AppName: rule.App.Name, AppBranchName: rule.AppBranch.Name}
	if rule.RunbookID == nil {
		return response, nil
	}
	var runbook app.Runbook
	if err := s.db.WithContext(ctx).Select("name").Where(app.Runbook{ID: *rule.RunbookID, OrgID: rule.OrgID}).First(&runbook).Error; err != nil {
		return response, err
	}
	response.RunbookName = runbook.Name
	return response, nil
}

// @ID ListTriggerRules
// @Summary List active rules for an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {array} triggerRuleResponse
// @Router /v1/triggers/{trigger_id}/rules [get]
func (s *service) ListTriggerRules(ctx *gin.Context) {
	trigger, ok := s.readTrigger(ctx)
	if !ok {
		return
	}
	configIDs, err := s.activeTriggerConfigIDs(ctx, trigger.OrgID)
	if err != nil {
		ctx.Error(err)
		return
	}
	var rules []app.TriggerRule
	if err := activeTriggerRulesQuery(s.db.WithContext(ctx), trigger, configIDs).Preload("App").Preload("AppBranch").Order("name ASC, id ASC").Find(&rules).Error; err != nil {
		ctx.Error(err)
		return
	}
	responses := make([]triggerRuleResponse, len(rules))
	for i := range rules {
		responses[i], err = s.triggerRuleResponse(ctx, rules[i])
		if err != nil {
			ctx.Error(err)
			return
		}
	}
	ctx.JSON(http.StatusOK, responses)
}

// @ID GetTriggerRule
// @Summary Get an active rule for an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Param rule_id path string true "Rule ID"
// @Success 200 {object} triggerRuleResponse
// @Router /v1/triggers/{trigger_id}/rules/{rule_id} [get]
func (s *service) GetTriggerRule(ctx *gin.Context) {
	trigger, ok := s.readTrigger(ctx)
	if !ok {
		return
	}
	configIDs, err := s.activeTriggerConfigIDs(ctx, trigger.OrgID)
	if err != nil {
		ctx.Error(err)
		return
	}
	var rule app.TriggerRule
	err = activeTriggerRulesQuery(s.db.WithContext(ctx), trigger, configIDs).Where(app.TriggerRule{ID: ctx.Param("rule_id")}).Preload("App").Preload("AppBranch").First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "trigger rule not found"})
		} else {
			ctx.Error(err)
		}
		return
	}
	response, err := s.triggerRuleResponse(ctx, rule)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}
