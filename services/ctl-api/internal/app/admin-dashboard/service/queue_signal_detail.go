package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) QueueSignalDetail(c *gin.Context) {
	ctx := c.Request.Context()
	queueID := c.Param("id")
	signalID := c.Param("signal_id")

	var signal app.QueueSignal
	res := s.db.WithContext(ctx).
		Preload("Emitter").
		Where("id = ? AND queue_id = ?", signalID, queueID).
		First(&signal)

	if res.Error != nil {
		s.l.Error("failed to fetch queue signal", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Signal not found"})
		return
	}

	var q app.Queue
	s.db.WithContext(ctx).
		Where("id = ?", queueID).
		First(&q)

	var wfInfo *views.WorkflowInfo
	if signal.Workflow.Namespace != "" && signal.Workflow.ID != "" {
		wfInfo = s.getWorkflowInfo(c, signal.Workflow.Namespace, signal.Workflow.ID)
		if wfInfo != nil {
			wfInfo.UpdateHandlers = updateHandlersForSignalType(string(signal.Type))
		}
	}

	attrs := signalAttributesForType(signal.Type)

	component := views.QueueSignalDetail(&signal, &q, s.cfg.TemporalUIURL, wfInfo, attrs)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

type scheduledActivity struct {
	name           string
	scheduledAt    time.Time
	scheduledEvtID int64
	input          string // JSON-formatted activity input
}

// updateState tracks an in-progress or completed workflow update.
type updateState struct {
	name          string
	updateID      string
	acceptedEvtID int64
	startedAt     time.Time
	finishedAt    time.Time
	status        string
	input         string
	result        string
	failure       string
}

func (s *service) getWorkflowInfo(c *gin.Context, namespace, workflowID string) *views.WorkflowInfo {
	ctx := c.Request.Context()

	// Get workflow status
	status, err := s.temporalClient.GetWorkflowStatusInNamespace(ctx, namespace, workflowID, "")
	if err != nil {
		s.l.Warn("failed to get workflow status", zap.Error(err), zap.String("workflow_id", workflowID))
		return nil
	}

	info := &views.WorkflowInfo{
		Status: formatWorkflowStatus(status),
	}

	// Get workflow history for activity details
	nsClient, err := s.temporalClient.GetNamespaceClient(namespace)
	if err != nil {
		s.l.Warn("failed to get namespace client", zap.Error(err))
		return info
	}

	iter := nsClient.GetWorkflowHistory(ctx, workflowID, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	scheduled := map[int64]scheduledActivity{}
	started := map[int64]time.Time{} // keyed by scheduled event ID
	attempts := map[int64]int32{}    // keyed by scheduled event ID

	// Track child workflows by initiated event ID
	type childWFState struct {
		workflowType string
		workflowID   string
		runID        string
		namespace    string
		startedAt    time.Time
	}
	childWFs := map[int64]childWFState{} // keyed by initiated event ID

	// Track update executions by accepted event ID
	updates := map[int64]*updateState{}

	var activities []views.ActivityInfo
	var childWorkflows []views.ChildWorkflowInfo

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			s.l.Warn("failed to iterate workflow history", zap.Error(err))
			break
		}

		switch event.GetEventType() {
		// Workflow Update tracking
		case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_UPDATE_ACCEPTED:
			attrs := event.GetWorkflowExecutionUpdateAcceptedEventAttributes()
			if attrs != nil {
				us := &updateState{
					acceptedEvtID: event.GetEventId(),
					startedAt:     event.GetEventTime().AsTime(),
					status:        "Running",
				}
				if req := attrs.GetAcceptedRequest(); req != nil {
					if meta := req.GetMeta(); meta != nil {
						us.updateID = meta.GetUpdateId()
					}
					if input := req.GetInput(); input != nil {
						us.name = input.GetName()
						us.input = s.formatPayloads(input.GetArgs())
					}
				}
				updates[event.GetEventId()] = us
			}

		case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_UPDATE_COMPLETED:
			attrs := event.GetWorkflowExecutionUpdateCompletedEventAttributes()
			if attrs != nil {
				if us, ok := updates[attrs.GetAcceptedEventId()]; ok {
					us.finishedAt = event.GetEventTime().AsTime()
					us.status = "Completed"
					if outcome := attrs.GetOutcome(); outcome != nil {
						if f := outcome.GetFailure(); f != nil {
							us.status = "Failed"
							us.failure = f.GetMessage()
						} else if successPayloads := outcome.GetSuccess(); successPayloads != nil {
							us.result = s.formatPayloads(successPayloads)
						}
					}
				}
			}

		case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_UPDATE_REJECTED:
			// Rejected updates don't have an accepted event, skip

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			attrs := event.GetActivityTaskScheduledEventAttributes()
			if attrs != nil {
				name := ""
				if attrs.GetActivityType() != nil {
					name = attrs.GetActivityType().GetName()
				}
				scheduled[event.GetEventId()] = scheduledActivity{
					name:           name,
					scheduledAt:    event.GetEventTime().AsTime(),
					scheduledEvtID: event.GetEventId(),
					input:          s.formatPayloads(attrs.GetInput()),
				}
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED:
			attrs := event.GetActivityTaskStartedEventAttributes()
			if attrs != nil {
				started[attrs.GetScheduledEventId()] = event.GetEventTime().AsTime()
				attempts[attrs.GetScheduledEventId()] = attrs.GetAttempt()
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
			attrs := event.GetActivityTaskCompletedEventAttributes()
			if attrs != nil {
				ai := s.buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Completed", "",
				)
				ai.Result = s.formatPayloads(attrs.GetResult())
				activities = append(activities, ai)
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_FAILED:
			attrs := event.GetActivityTaskFailedEventAttributes()
			if attrs != nil {
				failure := ""
				if attrs.GetFailure() != nil {
					failure = attrs.GetFailure().GetMessage()
				}
				activities = append(activities, s.buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Failed", failure,
				))
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
			attrs := event.GetActivityTaskTimedOutEventAttributes()
			if attrs != nil {
				activities = append(activities, s.buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Timed Out", "",
				))
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
			attrs := event.GetActivityTaskCanceledEventAttributes()
			if attrs != nil {
				activities = append(activities, s.buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Canceled", "",
				))
			}

		// Child workflow tracking
		case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
			attrs := event.GetChildWorkflowExecutionStartedEventAttributes()
			if attrs != nil {
				wfType := ""
				if attrs.GetWorkflowType() != nil {
					wfType = attrs.GetWorkflowType().GetName()
				}
				wfExec := attrs.GetWorkflowExecution()
				state := childWFState{
					workflowType: wfType,
					startedAt:    event.GetEventTime().AsTime(),
				}
				if wfExec != nil {
					state.workflowID = wfExec.GetWorkflowId()
					state.runID = wfExec.GetRunId()
				}
				if attrs.GetNamespace() != "" {
					state.namespace = attrs.GetNamespace()
				}
				childWFs[attrs.GetInitiatedEventId()] = state
			}

		case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
			attrs := event.GetChildWorkflowExecutionCompletedEventAttributes()
			if attrs != nil {
				if state, ok := childWFs[attrs.GetInitiatedEventId()]; ok {
					childWorkflows = append(childWorkflows, views.ChildWorkflowInfo{
						WorkflowType: state.workflowType,
						WorkflowID:   state.workflowID,
						RunID:        state.runID,
						Namespace:    state.namespace,
						Status:       "Completed",
						StartedAt:    state.startedAt,
						FinishedAt:   event.GetEventTime().AsTime(),
						Duration:     event.GetEventTime().AsTime().Sub(state.startedAt),
					})
					delete(childWFs, attrs.GetInitiatedEventId())
				}
			}

		case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
			attrs := event.GetChildWorkflowExecutionFailedEventAttributes()
			if attrs != nil {
				if state, ok := childWFs[attrs.GetInitiatedEventId()]; ok {
					failure := ""
					if attrs.GetFailure() != nil {
						failure = attrs.GetFailure().GetMessage()
					}
					childWorkflows = append(childWorkflows, views.ChildWorkflowInfo{
						WorkflowType: state.workflowType,
						WorkflowID:   state.workflowID,
						RunID:        state.runID,
						Namespace:    state.namespace,
						Status:       "Failed",
						StartedAt:    state.startedAt,
						FinishedAt:   event.GetEventTime().AsTime(),
						Duration:     event.GetEventTime().AsTime().Sub(state.startedAt),
						Failure:      failure,
					})
					delete(childWFs, attrs.GetInitiatedEventId())
				}
			}

		case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
			attrs := event.GetChildWorkflowExecutionCanceledEventAttributes()
			if attrs != nil {
				if state, ok := childWFs[attrs.GetInitiatedEventId()]; ok {
					childWorkflows = append(childWorkflows, views.ChildWorkflowInfo{
						WorkflowType: state.workflowType,
						WorkflowID:   state.workflowID,
						RunID:        state.runID,
						Namespace:    state.namespace,
						Status:       "Canceled",
						StartedAt:    state.startedAt,
						FinishedAt:   event.GetEventTime().AsTime(),
						Duration:     event.GetEventTime().AsTime().Sub(state.startedAt),
					})
					delete(childWFs, attrs.GetInitiatedEventId())
				}
			}

		case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
			attrs := event.GetChildWorkflowExecutionTimedOutEventAttributes()
			if attrs != nil {
				if state, ok := childWFs[attrs.GetInitiatedEventId()]; ok {
					childWorkflows = append(childWorkflows, views.ChildWorkflowInfo{
						WorkflowType: state.workflowType,
						WorkflowID:   state.workflowID,
						RunID:        state.runID,
						Namespace:    state.namespace,
						Status:       "Timed Out",
						StartedAt:    state.startedAt,
						FinishedAt:   event.GetEventTime().AsTime(),
						Duration:     event.GetEventTime().AsTime().Sub(state.startedAt),
					})
					delete(childWFs, attrs.GetInitiatedEventId())
				}
			}
		}
	}

	// Add still-running child workflows
	for _, state := range childWFs {
		childWorkflows = append(childWorkflows, views.ChildWorkflowInfo{
			WorkflowType: state.workflowType,
			WorkflowID:   state.workflowID,
			RunID:        state.runID,
			Namespace:    state.namespace,
			Status:       "Running",
			StartedAt:    state.startedAt,
		})
	}

	// Add any scheduled-but-not-finished activities as "Running" or "Scheduled"
	for schedID, sched := range scheduled {
		found := false
		for _, a := range activities {
			if a.ScheduledEventID == schedID {
				found = true
				break
			}
		}
		if !found {
			actInfo := views.ActivityInfo{
				Name:             sched.name,
				ScheduledEventID: schedID,
			}
			if startTime, ok := started[schedID]; ok {
				actInfo.Status = "Running"
				actInfo.StartedAt = startTime
				if att, ok := attempts[schedID]; ok {
					actInfo.Attempt = att
				}
			} else {
				actInfo.Status = "Scheduled"
			}
			activities = append(activities, actInfo)
		}
	}

	info.Activities = activities
	info.ChildWorkflows = childWorkflows

	// Extract awaited signals from AwaitSignal activities
	info.AwaitedSignals = s.extractAwaitedSignals(c, activities)

	// Build update executions by associating activities with updates based on event ID ranges
	if len(updates) > 0 {
		info.UpdateExecutions, info.OrphanActivities = s.buildUpdateExecutions(updates, activities)
	} else {
		info.OrphanActivities = activities
	}

	return info
}

// extractAwaitedSignals looks for AwaitSignal activities in the workflow history
// and loads the corresponding queue signals from the database.
func (s *service) extractAwaitedSignals(c *gin.Context, activities []views.ActivityInfo) []views.AwaitedSignalInfo {
	ctx := c.Request.Context()
	var awaited []views.AwaitedSignalInfo

	for _, act := range activities {
		if !strings.Contains(act.Name, "AwaitSignal") {
			continue
		}

		// Extract queue signal ID from the activity input JSON
		// The input is the first argument: a string (queue signal ID)
		qsID := extractQueueSignalIDFromInput(act.Input)
		if qsID == "" {
			continue
		}

		asi := views.AwaitedSignalInfo{
			QueueSignalID: qsID,
			Status:        act.Status,
			StartedAt:     act.StartedAt,
			FinishedAt:    act.FinishedAt,
			Duration:      act.Duration,
			Failure:       act.Failure,
		}

		// Load the awaited signal from DB
		var signal app.QueueSignal
		if err := s.db.WithContext(ctx).Where("id = ?", qsID).First(&signal).Error; err == nil {
			asi.Signal = &signal
		}

		awaited = append(awaited, asi)
	}

	return awaited
}

// extractQueueSignalIDFromInput parses the activity input JSON to find a queue signal ID.
// The input format is a JSON string like "\"qsi...\""
func extractQueueSignalIDFromInput(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Try parsing as a plain JSON string (most common: AwaitSignal takes a single string arg)
	var id string
	if err := json.Unmarshal([]byte(input), &id); err == nil {
		if strings.HasPrefix(id, "qsi") {
			return id
		}
	}

	return ""
}

// buildUpdateExecutions groups activities into their parent update executions.
// Activities whose ScheduledEventID falls after an update's accepted event ID
// (and before the next update's) belong to that update.
// Activities before any update are returned as orphans (main workflow body).
func (s *service) buildUpdateExecutions(updates map[int64]*updateState, activities []views.ActivityInfo) ([]views.UpdateExecution, []views.ActivityInfo) {
	// Collect and sort updates by accepted event ID
	type indexedUpdate struct {
		acceptedEvtID int64
		state         *updateState
	}
	sorted := make([]indexedUpdate, 0, len(updates))
	for id, us := range updates {
		sorted = append(sorted, indexedUpdate{acceptedEvtID: id, state: us})
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].acceptedEvtID < sorted[i].acceptedEvtID {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// For each activity, find which update it belongs to
	actsByUpdate := make(map[int64][]views.ActivityInfo) // keyed by accepted event ID
	var orphans []views.ActivityInfo

	for _, act := range activities {
		ownerIdx := -1
		for i, su := range sorted {
			if act.ScheduledEventID <= su.acceptedEvtID {
				break // past all possible owners
			}
			// Activity scheduled after this update's accepted event
			if i+1 < len(sorted) && act.ScheduledEventID >= sorted[i+1].acceptedEvtID {
				continue // belongs to a later update
			}
			ownerIdx = i
			break
		}
		if ownerIdx >= 0 {
			key := sorted[ownerIdx].acceptedEvtID
			actsByUpdate[key] = append(actsByUpdate[key], act)
		} else {
			orphans = append(orphans, act)
		}
	}

	// Build UpdateExecution structs
	execs := make([]views.UpdateExecution, 0, len(sorted))
	for _, su := range sorted {
		ue := views.UpdateExecution{
			Name:       su.state.name,
			UpdateID:   su.state.updateID,
			Status:     su.state.status,
			StartedAt:  su.state.startedAt,
			Input:      su.state.input,
			Result:     su.state.result,
			Failure:    su.state.failure,
			Activities: actsByUpdate[su.acceptedEvtID],
		}
		if !su.state.finishedAt.IsZero() {
			ue.FinishedAt = su.state.finishedAt
			ue.Duration = su.state.finishedAt.Sub(su.state.startedAt)
		}
		execs = append(execs, ue)
	}

	return execs, orphans
}

func (s *service) buildActivityInfo(
	scheduled map[int64]scheduledActivity,
	started map[int64]time.Time,
	attempts map[int64]int32,
	scheduledEventID int64,
	event *historypb.HistoryEvent,
	status string,
	failure string,
) views.ActivityInfo {
	ai := views.ActivityInfo{
		Status:           status,
		FinishedAt:       event.GetEventTime().AsTime(),
		Failure:          failure,
		ScheduledEventID: scheduledEventID,
	}

	if sched, ok := scheduled[scheduledEventID]; ok {
		ai.Name = sched.name
		ai.Input = sched.input
	}
	if startTime, ok := started[scheduledEventID]; ok {
		ai.StartedAt = startTime
		ai.Duration = ai.FinishedAt.Sub(startTime)
	}
	if att, ok := attempts[scheduledEventID]; ok {
		ai.Attempt = att
	}

	return ai
}

// decodePayloads runs the codec chain to decode Temporal payloads (e.g. gzip, large payload, s3).
func (s *service) decodePayloads(payloads *commonpb.Payloads) *commonpb.Payloads {
	if payloads == nil || len(payloads.GetPayloads()) == 0 {
		return payloads
	}
	decoded := payloads.GetPayloads()
	for _, codec := range s.codecs {
		out, err := codec.Decode(decoded)
		if err != nil {
			s.l.Debug("codec decode failed, using raw payload", zap.Error(err))
			return payloads
		}
		decoded = out
	}
	return &commonpb.Payloads{Payloads: decoded}
}

// formatPayloads decodes and pretty-prints the JSON data from Temporal payloads.
func (s *service) formatPayloads(payloads *commonpb.Payloads) string {
	if payloads == nil {
		return ""
	}
	payloads = s.decodePayloads(payloads)

	var parts []json.RawMessage
	for _, p := range payloads.GetPayloads() {
		if p == nil || len(p.GetData()) == 0 {
			continue
		}
		if json.Valid(p.GetData()) {
			parts = append(parts, json.RawMessage(p.GetData()))
		} else {
			quoted, _ := json.Marshal(string(p.GetData()))
			parts = append(parts, json.RawMessage(quoted))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	var raw []byte
	if len(parts) == 1 {
		raw = parts[0]
	} else {
		raw, _ = json.Marshal(parts)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

// updateHandlersForSignalType returns the known update handler names registered
// by each signal type. Derived from the RegisterUpdateHandlers implementations.
func updateHandlersForSignalType(signalType string) []string {
	switch signalType {
	case "execute-flow":
		return []string{"retry-step", "approve-step", "skip-step", "cancel-step", "is-retryable", "poll-next-step"}
	case "execute-workflow-step":
		return []string{"is-retryable", "create-step-retry", "approve-plan"}
	default:
		return nil
	}
}

func formatWorkflowStatus(status enumspb.WorkflowExecutionStatus) string {
	switch status {
	case enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return "Running"
	case enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return "Completed"
	case enumspb.WORKFLOW_EXECUTION_STATUS_FAILED:
		return "Failed"
	case enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return "Canceled"
	case enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return "Terminated"
	case enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return "Continued As New"
	case enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return "Timed Out"
	default:
		return "Unknown"
	}
}
