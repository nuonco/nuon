package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID							GetWorkflowSteps
// @Summary						get all of the steps for a given workflow
// @Description.markdown		get_workflow_steps.md
// @Param workflow_id	path	string true "workflow ID"
// @Param				offset	query	int		false	"offset of results to return"	Default(0)
// @Param				limit	query	int		false	"limit of results to return"	Default(10)
// @Param				page	query	int		false	"page number of results to return"	Default(0)
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success						200	{array}		app.WorkflowStep
// @Router						/v1/workflows/{workflow_id}/steps [GET]
func (s *service) GetWorkflowSteps(ctx *gin.Context) {
	workflowID := ctx.Param("workflow_id")

	steps, err := s.getWorkflowSteps(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get workflow steps"))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

// TODO: Remove. Deprecated.
// @ID							GetInstallWorkflowSteps
// @Summary						get all of the steps for a given install workflow
// @Description.markdown		get_workflow_steps.md
// @Param install_workflow_id	path	string true "install workflow ID"
// @Param				offset	query	int		false	"offset of results to return"	Default(0)
// @Param				limit	query	int		false	"limit of results to return"	Default(10)
// @Param				page	query	int		false	"page number of results to return"	Default(0)
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						404	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success						200	{array}		app.WorkflowStep
// @Router						/v1/install-workflows/{install_workflow_id}/steps [GET]
// @Deprecated
func (s *service) GetInstallWorkflowSteps(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	steps, err := s.getWorkflowSteps(ctx, workflowID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install workflow steps"))
		return
	}

	ctx.JSON(http.StatusOK, steps)
}

func (s *service) getWorkflowSteps(ctx *gin.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.WorkflowStep{
			InstallWorkflowID: workflowID,
		}).
		Preload("CreatedBy").
		Preload("Approval").
		Preload("Approval.Response").
		Preload("PolicyValidation").
		Order("group_idx, group_retry_idx, idx, created_at asc").
		Find(&steps)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get workflow steps")
	}

	steps, err := db.HandlePaginatedResponse(ctx, steps)
	if err != nil {
		return nil, errors.Wrap(err, "unable to handle paginated response")
	}

	stepPtrs := make([]*app.WorkflowStep, len(steps))
	for i := range steps {
		stepPtrs[i] = &steps[i]
	}
	s.loadStepLogStreams(ctx, stepPtrs)
	if err := s.loadStepEventWaits(ctx, stepPtrs); err != nil {
		return nil, errors.Wrap(err, "unable to load workflow event waits")
	}

	return steps, nil
}

func (s *service) loadStepEventWaits(ctx *gin.Context, steps []*app.WorkflowStep) error {
	stepIDs := make([]string, len(steps))
	for i := range steps {
		stepIDs[i] = steps[i].ID
	}
	if len(stepIDs) == 0 {
		return nil
	}

	var waiters []app.EventRunbookWaiter
	if err := s.db.WithContext(ctx).Where(map[string]any{"workflow_step_id": stepIDs}).Find(&waiters).Error; err != nil {
		return err
	}
	waitersByStep := make(map[string]app.EventRunbookWaiter, len(waiters))
	triggerIDs := make([]string, 0, len(waiters))
	eventIDs := make([]string, 0, len(waiters))
	for i := range waiters {
		waiter := waiters[i]
		waitersByStep[waiter.WorkflowStepID] = waiter
		triggerIDs = append(triggerIDs, waiter.TriggerID)
		if waiter.MatchedEventID != nil {
			eventIDs = append(eventIDs, *waiter.MatchedEventID)
		}
	}

	var triggers []app.Trigger
	if len(triggerIDs) > 0 {
		if err := s.db.WithContext(ctx).Unscoped().Select("id", "name").Where(map[string]any{"id": triggerIDs}).Find(&triggers).Error; err != nil {
			return err
		}
	}
	triggerNames := make(map[string]string, len(triggers))
	for i := range triggers {
		triggerNames[triggers[i].ID] = triggers[i].Name
	}

	var events []app.TriggerEvent
	if len(eventIDs) > 0 {
		if err := s.db.WithContext(ctx).Select("id", "event_type").Where(map[string]any{"id": eventIDs}).Find(&events).Error; err != nil {
			return err
		}
	}
	eventTypes := make(map[string]string, len(events))
	for i := range events {
		eventTypes[events[i].ID] = events[i].EventType
	}

	for i := range steps {
		step := steps[i]
		waiter, ok := waitersByStep[step.ID]
		if !ok {
			continue
		}
		if step.Links == nil {
			step.Links = make(map[string]any)
		}
		matchedEventType := ""
		if waiter.MatchedEventID != nil {
			matchedEventType = eventTypes[*waiter.MatchedEventID]
		}
		step.Links["event_wait"] = map[string]any{
			"id":                 waiter.ID,
			"status":             waiter.Status,
			"trigger_id":         waiter.TriggerID,
			"trigger_name":       triggerNames[waiter.TriggerID],
			"event_types":        waiter.EventTypes,
			"filters":            waiter.Filters,
			"matched_event_id":   waiter.MatchedEventID,
			"matched_event_type": matchedEventType,
			"activated_at":       waiter.ActivatedAt,
			"matched_at":         waiter.MatchedAt,
			"notified_at":        waiter.NotifiedAt,
			"cancelled_at":       waiter.CancelledAt,
			"expired_at":         waiter.ExpiredAt,
		}
	}
	return nil
}
