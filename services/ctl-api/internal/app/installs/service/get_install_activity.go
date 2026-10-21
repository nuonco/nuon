package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallActivityType string

const (
	InstallActivityTypeActionRun   InstallActivityType = "action_run"
	InstallActivityTypeRunbookRun  InstallActivityType = "runbook_run"
	InstallActivityTypePolicyCheck InstallActivityType = "policy_check"
)

type InstallActivityWorkflowRef struct {
	ID   string           `json:"id"`
	Type app.WorkflowType `json:"type"`
	Name string           `json:"name"`
}

type InstallActivityAction struct {
	RunID            string                `json:"run_id"`
	ActionWorkflowID string                `json:"action_workflow_id,omitempty"`
	Name             string                `json:"name,omitempty"`
	TriggerType      string                `json:"trigger_type,omitempty"`
	Trigger          *app.ActionRunTrigger `json:"trigger,omitempty"`
	Adhoc            bool                  `json:"adhoc,omitempty"`
}

type InstallActivityRunbook struct {
	RunID     string `json:"run_id"`
	RunbookID string `json:"runbook_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

type InstallActivityPolicy struct {
	ReportID      string `json:"report_id"`
	OwnerType     string `json:"owner_type"`
	OwnerID       string `json:"owner_id"`
	ComponentName string `json:"component_name,omitempty"`
	DenyCount     int    `json:"deny_count"`
	WarnCount     int    `json:"warn_count"`
	PassCount     int    `json:"pass_count"`
}

type InstallActivity struct {
	ID        string              `json:"id"`
	Type      InstallActivityType `json:"type"`
	Status    string              `json:"status"`
	CreatedAt time.Time           `json:"created_at"`
	Title     string              `json:"title"`
	Summary   string              `json:"summary"`
	Actor     *app.Account        `json:"actor,omitempty"`

	actorID string

	Workflow *InstallActivityWorkflowRef `json:"workflow,omitempty"`
	Action   *InstallActivityAction      `json:"action,omitempty"`
	Runbook  *InstallActivityRunbook     `json:"runbook,omitempty"`
	Policy   *InstallActivityPolicy      `json:"policy,omitempty"`
}

type GetInstallActivityResponse struct {
	Activity   []InstallActivity `json:"activity"`
	Page       int               `json:"page"`
	Offset     int               `json:"offset"`
	Limit      int               `json:"limit"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

type installActivityCursor struct {
	CreatedAt time.Time           `json:"t"`
	Type      InstallActivityType `json:"k"`
	ID        string              `json:"i"`
}

func decodeInstallActivityCursor(raw string) (*installActivityCursor, error) {
	if len(raw) > 1024 {
		return nil, errors.New("cursor is not valid")
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, errors.New("cursor is not valid")
	}
	var cursor installActivityCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, errors.New("cursor is not valid")
	}
	if cursor.CreatedAt.IsZero() || len(cursor.ID) != 26 {
		return nil, errors.New("cursor is not valid")
	}
	switch cursor.Type {
	case InstallActivityTypeActionRun, InstallActivityTypeRunbookRun, InstallActivityTypePolicyCheck:
		return &cursor, nil
	default:
		return nil, errors.New("cursor is not valid")
	}
}

// @ID                    GetInstallActivity
// @Summary               get normalized activity feed for an install
// @Description.markdown  get_install_activity.md
// @Param                 install_id      path   string  true   "install ID"
// @Param                 page            query  int     false  "page number"                                         Default(0)
// @Param                 offset          query  int     false  "offset of results to return"                         Default(0)
// @Param                 limit           query  int     false  "page size"                                           Default(20)
// @Param                 cursor          query  string  false  "opaque cursor from next_cursor; replaces page and offset"
// @Param                 type            query  string  false  "filter by activity type (comma-separated: action_run, runbook_run, policy_check)"
// @Param                 status          query  string  false  "filter by source status (comma-separated)"
// @Param                 search          query  string  false  "case-insensitive substring match on id or title"
// @Param                 actor_id        query  string  false  "only include activity triggered by this account ID"
// @Param                 exclude_cron    query  bool    false  "exclude cron-triggered action and runbook runs"
// @Param                 created_at_gte  query  string  false  "include activity created at or after this RFC3339 timestamp"
// @Param                 created_at_lte  query  string  false  "include activity created at or before this RFC3339 timestamp"
// @Tags                  installs
// @Accept                json
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400  {object}  stderr.ErrResponse
// @Failure               401  {object}  stderr.ErrResponse
// @Failure               403  {object}  stderr.ErrResponse
// @Failure               404  {object}  stderr.ErrResponse
// @Failure               500  {object}  stderr.ErrResponse
// @Success               200  {object}  GetInstallActivityResponse
// @Router                /v1/installs/{install_id}/activity [GET]
func (s *service) GetInstallActivity(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	limit := queryInt(ctx, "limit", 20, 1, 100)
	page := queryInt(ctx, "page", 0, 0, 10_000)
	offset := queryInt(ctx, "offset", 0, 0, 1_000_000)
	if ctx.Query("page") != "" {
		offset = page * limit
	} else {
		page = offset / limit
	}

	var cursor *installActivityCursor
	if raw := ctx.Query("cursor"); raw != "" {
		if offset > 0 {
			ctx.Error(stderr.ErrUser{
				Err:         errors.New("cursor cannot be combined with page or offset"),
				Description: "use either cursor or page/offset",
			})
			return
		}
		cursor, err = decodeInstallActivityCursor(raw)
		if err != nil {
			ctx.Error(stderr.ErrUser{Err: err, Description: "invalid cursor"})
			return
		}
	}

	var createdAtGte *time.Time
	if raw := ctx.Query("created_at_gte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_gte, must be RFC3339"))
			return
		}
		createdAtGte = &t
	}

	var createdAtLte *time.Time
	if raw := ctx.Query("created_at_lte"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			ctx.Error(errors.Wrap(err, "invalid created_at_lte, must be RFC3339"))
			return
		}
		createdAtLte = &t
	}

	resp, err := s.getInstallActivity(
		ctx,
		org.ID,
		installID,
		page,
		offset,
		limit,
		parseCommaSeparated(ctx.Query("type")),
		parseCommaSeparated(ctx.Query("status")),
		ctx.Query("search"),
		ctx.Query("actor_id"),
		ctx.Query("exclude_cron") == "true",
		createdAtGte,
		createdAtLte,
		cursor,
	)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install activity: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getInstallActivity(
	ctx *gin.Context,
	orgID, installID string,
	page, offset, limit int,
	filterTypes, filterStatuses []string,
	search, actorID string,
	excludeCron bool,
	createdAtGte, createdAtLte *time.Time,
	cursor *installActivityCursor,
) (*GetInstallActivityResponse, error) {
	selected := selectedActivityTypes(filterTypes)
	empty := &GetInstallActivityResponse{
		Activity: []InstallActivity{},
		Page:     page,
		Offset:   offset,
		Limit:    limit,
	}
	if actorID != "" {
		delete(selected, InstallActivityTypePolicyCheck)
	}
	if len(selected) == 0 {
		return empty, nil
	}

	fetchLimit := offset + limit + 1
	activity := make([]InstallActivity, 0)

	if selected[InstallActivityTypeActionRun] {
		runs, err := s.listActivityActionRuns(ctx, orgID, installID, fetchLimit, filterStatuses, search, actorID, excludeCron, createdAtGte, createdAtLte, cursor)
		if err != nil {
			return nil, err
		}
		for i := range runs {
			activity = append(activity, actionRunActivity(&runs[i]))
		}
	}

	if selected[InstallActivityTypeRunbookRun] {
		runs, err := s.listActivityRunbookRuns(ctx, orgID, installID, fetchLimit, filterStatuses, search, actorID, excludeCron, createdAtGte, createdAtLte, cursor)
		if err != nil {
			return nil, err
		}
		for i := range runs {
			activity = append(activity, runbookRunActivity(&runs[i]))
		}
	}

	if selected[InstallActivityTypePolicyCheck] {
		reports, err := s.listActivityPolicyReports(ctx, orgID, installID, fetchLimit, filterStatuses, search, createdAtGte, createdAtLte, cursor)
		if err != nil {
			return nil, err
		}
		for i := range reports {
			activity = append(activity, policyReportActivity(&reports[i]))
		}
	}

	sort.SliceStable(activity, func(i, j int) bool {
		if activity[i].CreatedAt.Equal(activity[j].CreatedAt) {
			if activity[i].Type == activity[j].Type {
				return activity[i].ID > activity[j].ID
			}
			return activity[i].Type > activity[j].Type
		}
		return activity[i].CreatedAt.After(activity[j].CreatedAt)
	})

	start := offset
	if start > len(activity) {
		start = len(activity)
	}
	end := start + limit
	hasMore := end < len(activity)
	if end > len(activity) {
		end = len(activity)
	}

	var nextCursor string
	if hasMore && end > start {
		last := activity[end-1]
		data, err := json.Marshal(installActivityCursor{CreatedAt: last.CreatedAt, Type: last.Type, ID: last.ID})
		if err != nil {
			return nil, fmt.Errorf("unable to encode cursor: %w", err)
		}
		nextCursor = base64.RawURLEncoding.EncodeToString(data)
	}

	pageItems := activity[start:end]
	if err := s.attachActivityActors(ctx, pageItems); err != nil {
		return nil, err
	}
	if err := s.attachActivityAdhocNames(ctx, pageItems); err != nil {
		return nil, err
	}

	return &GetInstallActivityResponse{
		Activity:   pageItems,
		Page:       page,
		Offset:     offset,
		Limit:      limit,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

func (s *service) listActivityActionRuns(
	ctx *gin.Context,
	orgID, installID string,
	fetchLimit int,
	filterStatuses []string,
	search, actorID string,
	excludeCron bool,
	createdAtGte, createdAtLte *time.Time,
	cursor *installActivityCursor,
) ([]app.InstallActionWorkflowRun, error) {
	query := s.db.WithContext(ctx).
		Select("id", "created_at", "install_action_workflow_id", "install_workflow_id", "status", "status_description", "status_v2", "composite_error", "trigger_type", "triggered_by_id", "triggered_by_type", "created_by_id").
		Preload("InstallActionWorkflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "action_workflow_id")
		}).
		Preload("InstallActionWorkflow.ActionWorkflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Where(app.InstallActionWorkflowRun{OrgID: orgID, InstallID: installID}).
		Where(`install_workflow_id IS NULL OR EXISTS (
			SELECT 1
			FROM install_workflows
			WHERE install_workflows.id = install_action_workflow_runs.install_workflow_id
			  AND install_workflows.type = ?
		)`, app.WorkflowTypeActionWorkflowRun).
		Order("install_action_workflow_runs.created_at DESC, install_action_workflow_runs.id DESC").
		Limit(fetchLimit)
	query = applyActivityCursor(query, "install_action_workflow_runs.created_at", "install_action_workflow_runs.id", InstallActivityTypeActionRun, cursor)
	query = applyActivityStatusFilter(query, "install_action_workflow_runs.status", filterStatuses)
	query = applyActivityTimeFilter(query, "install_action_workflow_runs.created_at", createdAtGte, createdAtLte)
	if excludeCron || actorID != "" {
		query = query.Where("install_action_workflow_runs.trigger_type <> 'cron'")
	}
	if actorID != "" {
		query = query.
			Where(`CASE
				WHEN install_action_workflow_runs.triggered_by_type = ? AND install_action_workflow_runs.triggered_by_id <> ''
				THEN install_action_workflow_runs.triggered_by_id
				ELSE install_action_workflow_runs.created_by_id
			END = ?`, app.ActionWorkflowTriggerTypeManual, actorID)
	}
	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where(`install_action_workflow_runs.id ILIKE ?
			OR EXISTS (
				SELECT 1
				FROM install_action_workflows iaw
				JOIN action_workflows aw ON aw.id = iaw.action_workflow_id AND aw.deleted_at = 0
				WHERE iaw.id = install_action_workflow_runs.install_action_workflow_id
				  AND iaw.deleted_at = 0
				  AND aw.name ILIKE ?
			)
			OR EXISTS (
				SELECT 1
				FROM install_action_workflow_run_steps iaws
				WHERE iaws.install_action_workflow_run_id = install_action_workflow_runs.id
				  AND iaws.deleted_at = 0
				  AND iaws.ad_hoc_config ->> 'name' ILIKE ?
			)`, like, like, like)
	}

	var runs []app.InstallActionWorkflowRun
	if err := query.Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("unable to query action runs: %w", err)
	}
	return runs, nil
}

func (s *service) listActivityRunbookRuns(
	ctx *gin.Context,
	orgID, installID string,
	fetchLimit int,
	filterStatuses []string,
	search, actorID string,
	excludeCron bool,
	createdAtGte, createdAtLte *time.Time,
	cursor *installActivityCursor,
) ([]app.InstallRunbookRun, error) {
	query := s.db.WithContext(ctx).
		Select("id", "created_at", "install_runbook_id", "install_workflow_id", runbookRunActivityStatusExpression+" AS status", runbookRunActivityStatusDescriptionExpression+" AS status_description", "created_by_id", "idempotency_key").
		Preload("InstallRunbook", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "runbook_id")
		}).
		Preload("InstallRunbook.Runbook", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Where(app.InstallRunbookRun{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC, id DESC").
		Limit(fetchLimit)
	query = applyActivityCursor(query, "created_at", "id", InstallActivityTypeRunbookRun, cursor)
	query = applyActivityStatusFilter(query, runbookRunActivityStatusExpression, filterStatuses)
	query = applyActivityTimeFilter(query, "created_at", createdAtGte, createdAtLte)
	if excludeCron || actorID != "" {
		query = query.Where("install_runbook_runs.idempotency_key IS NULL OR install_runbook_runs.idempotency_key NOT LIKE ?", runbookCronIdempotencyPrefix+"%")
	}
	if actorID != "" {
		query = query.Where(app.InstallRunbookRun{CreatedByID: actorID})
	}
	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where(`install_runbook_runs.id ILIKE ? OR EXISTS (
			SELECT 1
			FROM install_runbooks ir
			JOIN runbooks rb ON rb.id = ir.runbook_id AND rb.deleted_at = 0
			WHERE ir.id = install_runbook_runs.install_runbook_id
			  AND ir.deleted_at = 0
			  AND rb.name ILIKE ?
		)`, like, like)
	}

	var runs []app.InstallRunbookRun
	if err := query.Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("unable to query runbook runs: %w", err)
	}
	return runs, nil
}

func (s *service) listActivityPolicyReports(
	ctx *gin.Context,
	orgID, installID string,
	fetchLimit int,
	filterStatuses []string,
	search string,
	createdAtGte, createdAtLte *time.Time,
	cursor *installActivityCursor,
) ([]app.PolicyReport, error) {
	query := s.db.WithContext(ctx).
		Select("id", "created_at", "evaluated_at", "status", "owner_type", "owner_id", "component_name", "deny_count", "warn_count", "pass_count").
		Where(app.PolicyReport{OrgID: orgID}).
		Where("install_id = ?", installID).
		Order(app.PolicyReportActivityTimeExpression + " DESC, id DESC").
		Limit(fetchLimit)
	query = applyActivityCursor(query, app.PolicyReportActivityTimeExpression, "id", InstallActivityTypePolicyCheck, cursor)
	if len(filterStatuses) > 0 {
		query = query.Where("status->>'status' IN ?", filterStatuses)
	}
	query = applyActivityTimeFilter(query, "evaluated_at", createdAtGte, createdAtLte)
	for _, token := range strings.Fields(search) {
		like := "%" + token + "%"
		query = query.Where("id ILIKE ? OR component_name ILIKE ?", like, like)
	}

	var reports []app.PolicyReport
	if err := query.Find(&reports).Error; err != nil {
		return nil, fmt.Errorf("unable to query policy reports: %w", err)
	}
	return reports, nil
}

const runbookCronIdempotencyPrefix = "runbook-cron:"

const runbookRunActivityStatusExpression = `COALESCE(
	NULLIF(install_runbook_runs.status_v2->>'status', ''),
	(
		SELECT CASE w.status->>'status' WHEN 'success' THEN 'finished' ELSE NULLIF(w.status->>'status', '') END
		FROM install_workflows w
		WHERE w.id = install_runbook_runs.install_workflow_id
	),
	install_runbook_runs.status
)`

const runbookRunActivityStatusDescriptionExpression = `COALESCE(
	NULLIF(install_runbook_runs.status_v2->>'status_human_description', ''),
	(
		SELECT NULLIF(w.status->>'status_human_description', '')
		FROM install_workflows w
		WHERE w.id = install_runbook_runs.install_workflow_id
	),
	install_runbook_runs.status_description
)`

func (s *service) attachActivityAdhocNames(ctx *gin.Context, activity []InstallActivity) error {
	runIDs := make([]string, 0)
	for _, item := range activity {
		if item.Action != nil && item.Action.Adhoc {
			runIDs = append(runIDs, item.Action.RunID)
		}
	}
	if len(runIDs) == 0 {
		return nil
	}

	var steps []app.InstallActionWorkflowRunStep
	if err := s.db.WithContext(ctx).
		Select("InstallActionWorkflowRunID", "AdHocConfig").
		Where("install_action_workflow_run_id IN ?", runIDs).
		Find(&steps).Error; err != nil {
		return fmt.Errorf("unable to get adhoc action names: %w", err)
	}
	names := make(map[string]string, len(steps))
	for _, step := range steps {
		if step.AdHocConfig != nil && step.AdHocConfig.Name != "" {
			names[step.InstallActionWorkflowRunID] = step.AdHocConfig.Name
		}
	}

	for i := range activity {
		action := activity[i].Action
		if action == nil || !action.Adhoc {
			continue
		}
		if name, ok := names[action.RunID]; ok {
			action.Name = name
			activity[i].Title = name
		}
	}
	return nil
}

func (s *service) attachActivityActors(ctx *gin.Context, activity []InstallActivity) error {
	ids := make([]string, 0, len(activity))
	seen := make(map[string]struct{}, len(activity))
	for _, item := range activity {
		if item.actorID == "" {
			continue
		}
		if _, ok := seen[item.actorID]; ok {
			continue
		}
		seen[item.actorID] = struct{}{}
		ids = append(ids, item.actorID)
	}
	if len(ids) == 0 {
		return nil
	}

	var accounts []app.Account
	if err := s.db.WithContext(ctx).
		Select("id", "email", "name", "account_type").
		Where("id IN ?", ids).
		Find(&accounts).Error; err != nil {
		return fmt.Errorf("unable to get activity actors: %w", err)
	}
	byID := make(map[string]*app.Account, len(accounts))
	for i := range accounts {
		byID[accounts[i].ID] = &accounts[i]
	}

	for i := range activity {
		actor, ok := byID[activity[i].actorID]
		if !ok {
			continue
		}
		activity[i].Actor = actor
		if activity[i].Action != nil && activity[i].Action.Trigger != nil {
			activity[i].Action.Trigger.Actor = actor
		}
	}
	return nil
}

func applyActivityCursor(query *gorm.DB, timeColumn, idColumn string, activityType InstallActivityType, cursor *installActivityCursor) *gorm.DB {
	if cursor == nil {
		return query
	}
	if activityType == cursor.Type {
		return query.Where("("+timeColumn+", "+idColumn+") < (?, ?)", cursor.CreatedAt, cursor.ID)
	}
	if activityType < cursor.Type {
		return query.Where(timeColumn+" <= ?", cursor.CreatedAt)
	}
	return query.Where(timeColumn+" < ?", cursor.CreatedAt)
}

func applyActivityStatusFilter(query *gorm.DB, column string, statuses []string) *gorm.DB {
	if len(statuses) == 0 {
		return query
	}
	return query.Where(column+" IN ?", statuses)
}

func applyActivityTimeFilter(query *gorm.DB, column string, createdAtGte, createdAtLte *time.Time) *gorm.DB {
	if createdAtGte != nil {
		query = query.Where(column+" >= ?", createdAtGte)
	}
	if createdAtLte != nil {
		query = query.Where(column+" <= ?", createdAtLte)
	}
	return query
}

func selectedActivityTypes(filters []string) map[InstallActivityType]bool {
	known := []InstallActivityType{
		InstallActivityTypeActionRun,
		InstallActivityTypeRunbookRun,
		InstallActivityTypePolicyCheck,
	}
	selected := make(map[InstallActivityType]bool, len(known))
	if len(filters) == 0 {
		for _, activityType := range known {
			selected[activityType] = true
		}
		return selected
	}
	allowed := make(map[InstallActivityType]struct{}, len(known))
	for _, activityType := range known {
		allowed[activityType] = struct{}{}
	}
	for _, filter := range filters {
		activityType := InstallActivityType(filter)
		if _, ok := allowed[activityType]; ok {
			selected[activityType] = true
		}
	}
	return selected
}

func actionRunActivity(run *app.InstallActionWorkflowRun) InstallActivity {
	name := run.InstallActionWorkflow.ActionWorkflow.Name
	title := name
	if title == "" {
		title = "Action run"
	}
	summary := actionRunActivitySummary(run)
	if summary == "" {
		summary = string(run.TriggerType)
	}
	if summary == "" {
		summary = "Action run"
	}

	trigger := &app.ActionRunTrigger{Type: app.ActionRunTriggerTypeFor(run)}
	actorID := ""
	switch {
	case run.TriggerType == app.ActionWorkflowTriggerTypeCron:
		trigger.Type = app.ActionRunTriggerTypeCron
	case run.TriggeredByID != "" && run.TriggeredByType == string(app.ActionWorkflowTriggerTypeManual):
		actorID = run.TriggeredByID
	default:
		actorID = run.CreatedByID
	}

	item := InstallActivity{
		ID:        run.ID,
		Type:      InstallActivityTypeActionRun,
		Status:    string(run.Status),
		CreatedAt: run.CreatedAt,
		Title:     title,
		Summary:   summary,
		Workflow:  activityWorkflowRef(run.InstallWorkflowID, app.WorkflowTypeActionWorkflowRun, title),
		Action: &InstallActivityAction{
			RunID:            run.ID,
			ActionWorkflowID: run.InstallActionWorkflow.ActionWorkflowID,
			Name:             name,
			TriggerType:      string(run.TriggerType),
			Trigger:          trigger,
			Adhoc:            run.TriggerType == app.ActionWorkflowTriggerTypeAdHoc,
		},
		actorID: actorID,
	}
	return item
}

func actionRunActivitySummary(run *app.InstallActionWorkflowRun) string {
	if run.CompositeError != nil && run.CompositeError.Message != "" {
		return run.CompositeError.Message
	}

	description := run.StatusDescription
	normalized := strings.ToLower(description)
	if strings.Contains(normalized, "runner did not pick up the job within the available timeout") ||
		strings.Contains(normalized, "runner did not reserve it before the pickup timeout") ||
		(strings.Contains(normalized, "runner") &&
			strings.Contains(normalized, "pick up") &&
			strings.Contains(normalized, "timeout")) {
		return "Runner did not pick up the job before it timed out."
	}
	return description
}

func runbookRunActivity(run *app.InstallRunbookRun) InstallActivity {
	name := run.InstallRunbook.Runbook.Name
	title := name
	if title == "" {
		title = "Runbook run"
	}
	summary := run.StatusDescription
	if summary == "" {
		summary = "Runbook run"
	}

	return InstallActivity{
		ID:        run.ID,
		Type:      InstallActivityTypeRunbookRun,
		Status:    string(run.Status),
		CreatedAt: run.CreatedAt,
		Title:     title,
		Summary:   summary,
		Workflow:  activityWorkflowRef(run.InstallWorkflowID, app.WorkflowTypeRunbookRun, title),
		Runbook: &InstallActivityRunbook{
			RunID:     run.ID,
			RunbookID: run.InstallRunbook.RunbookID,
			Name:      name,
		},
		actorID: runbookRunActorID(run),
	}
}

func runbookRunActorID(run *app.InstallRunbookRun) string {
	if run.IdempotencyKey != nil && strings.HasPrefix(*run.IdempotencyKey, runbookCronIdempotencyPrefix) {
		return ""
	}
	return run.CreatedByID
}

func policyReportActivity(report *app.PolicyReport) InstallActivity {
	componentName := ""
	if report.ComponentName != nil {
		componentName = *report.ComponentName
	}
	title := componentName
	if title == "" {
		switch report.OwnerType {
		case app.PolicyReportOwnerTypeInstallSandboxRun:
			title = "Sandbox"
		default:
			title = "Policy check"
		}
	}
	summary := report.Status.StatusHumanDescription
	if summary == "" {
		summary = fmt.Sprintf("%d denied, %d warnings, %d passed", report.DenyCount, report.WarnCount, report.PassCount)
	}
	createdAt := report.EvaluatedAt
	if createdAt.IsZero() {
		createdAt = report.CreatedAt
	}
	status := string(report.Status.Status)
	if status == "" {
		status = string(app.StatusSuccess)
	}

	return InstallActivity{
		ID:        report.ID,
		Type:      InstallActivityTypePolicyCheck,
		Status:    status,
		CreatedAt: createdAt,
		Title:     title,
		Summary:   summary,
		Policy: &InstallActivityPolicy{
			ReportID:      report.ID,
			OwnerType:     string(report.OwnerType),
			OwnerID:       report.OwnerID,
			ComponentName: componentName,
			DenyCount:     report.DenyCount,
			WarnCount:     report.WarnCount,
			PassCount:     report.PassCount,
		},
	}
}

func activityWorkflowRef(id *string, workflowType app.WorkflowType, name string) *InstallActivityWorkflowRef {
	if id == nil || *id == "" {
		return nil
	}
	return &InstallActivityWorkflowRef{
		ID:   *id,
		Type: workflowType,
		Name: name,
	}
}
