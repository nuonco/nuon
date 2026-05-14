package helpers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Helpers provides shared graph-building and workflow info logic for signals.
type Helpers struct {
	db             *gorm.DB
	temporalClient temporalclient.Client
	l              *zap.Logger
}

// New creates a new helpers instance.
func New(db *gorm.DB, temporalClient temporalclient.Client, l *zap.Logger) *Helpers {
	return &Helpers{
		db:             db,
		temporalClient: temporalClient,
		l:              l,
	}
}

type scheduledActivity struct {
	name           string
	scheduledAt    time.Time
	scheduledEvtID int64
	input          string
}

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

// GetWorkflowInfo fetches Temporal workflow execution info for display.
// It does not decode payloads (no codecs) - input/result fields will be empty.
func (h *Helpers) GetWorkflowInfo(ctx context.Context, namespace, workflowID string) *WorkflowInfo {
	status, err := h.temporalClient.GetWorkflowStatusInNamespace(ctx, namespace, workflowID, "")
	if err != nil {
		h.l.Warn("failed to get workflow status", zap.Error(err), zap.String("workflow_id", workflowID))
		return nil
	}

	info := &WorkflowInfo{
		Status: FormatWorkflowStatus(status),
	}

	nsClient, err := h.temporalClient.GetNamespaceClient(namespace)
	if err != nil {
		h.l.Warn("failed to get namespace client", zap.Error(err))
		return info
	}

	iter := nsClient.GetWorkflowHistory(ctx, workflowID, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	scheduled := map[int64]scheduledActivity{}
	started := map[int64]time.Time{}
	attempts := map[int64]int32{}

	type childWFState struct {
		workflowType string
		workflowID   string
		runID        string
		namespace    string
		startedAt    time.Time
	}
	childWFs := map[int64]childWFState{}

	updates := map[int64]*updateState{}

	var activities []ActivityInfo
	var childWorkflows []ChildWorkflowInfo

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			h.l.Warn("failed to iterate workflow history", zap.Error(err))
			break
		}

		switch event.GetEventType() {
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
						}
					}
				}
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			attrs := event.GetActivityTaskScheduledEventAttributes()
			if attrs != nil {
				name := ""
				if attrs.GetActivityType() != nil {
					name = attrs.GetActivityType().GetName()
				}
				sa := scheduledActivity{
					name:           name,
					scheduledAt:    event.GetEventTime().AsTime(),
					scheduledEvtID: event.GetEventId(),
				}
				// Extract simple payload data for signal ID parsing
				if input := attrs.GetInput(); input != nil {
					sa.input = formatPayloadsSimple(input)
				}
				scheduled[event.GetEventId()] = sa
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
				ai := buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Completed", "",
				)
				if result := attrs.GetResult(); result != nil {
					ai.Result = formatPayloadsSimple(result)
				}
				activities = append(activities, ai)
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_FAILED:
			attrs := event.GetActivityTaskFailedEventAttributes()
			if attrs != nil {
				failure := ""
				if attrs.GetFailure() != nil {
					failure = attrs.GetFailure().GetMessage()
				}
				activities = append(activities, buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Failed", failure,
				))
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
			attrs := event.GetActivityTaskTimedOutEventAttributes()
			if attrs != nil {
				activities = append(activities, buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Timed Out", "",
				))
			}

		case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
			attrs := event.GetActivityTaskCanceledEventAttributes()
			if attrs != nil {
				activities = append(activities, buildActivityInfo(
					scheduled, started, attempts,
					attrs.GetScheduledEventId(),
					event, "Canceled", "",
				))
			}

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
					childWorkflows = append(childWorkflows, ChildWorkflowInfo{
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
					childWorkflows = append(childWorkflows, ChildWorkflowInfo{
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
					childWorkflows = append(childWorkflows, ChildWorkflowInfo{
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
					childWorkflows = append(childWorkflows, ChildWorkflowInfo{
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
		childWorkflows = append(childWorkflows, ChildWorkflowInfo{
			WorkflowType: state.workflowType,
			WorkflowID:   state.workflowID,
			RunID:        state.runID,
			Namespace:    state.namespace,
			Status:       "Running",
			StartedAt:    state.startedAt,
		})
	}

	// Add scheduled-but-not-finished activities
	for schedID, sched := range scheduled {
		found := false
		for _, a := range activities {
			if a.ScheduledEventID == schedID {
				found = true
				break
			}
		}
		if !found {
			actInfo := ActivityInfo{
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

	info.AwaitedSignals = h.ExtractAwaitedSignals(ctx, activities)
	info.EnqueuedSignals = h.ExtractEnqueuedSignals(ctx, activities)

	if len(updates) > 0 {
		info.UpdateExecutions, info.OrphanActivities = BuildUpdateExecutions(updates, activities, info.AwaitedSignals, info.EnqueuedSignals)
	} else {
		info.OrphanActivities = activities
	}

	return info
}

// ExtractAwaitedSignals looks for AwaitSignal activities and loads the corresponding queue signals.
func (h *Helpers) ExtractAwaitedSignals(ctx context.Context, activities []ActivityInfo) []AwaitedSignalInfo {
	var awaited []AwaitedSignalInfo

	for _, act := range activities {
		if !strings.Contains(act.Name, "AwaitSignal") {
			continue
		}

		qsID := ExtractQueueSignalIDFromInput(act.Input)
		if qsID == "" {
			continue
		}

		asi := AwaitedSignalInfo{
			QueueSignalID: qsID,
			Status:        act.Status,
			StartedAt:     act.StartedAt,
			FinishedAt:    act.FinishedAt,
			Duration:      act.Duration,
			Failure:       act.Failure,
		}

		var signal app.QueueSignal
		if err := h.db.WithContext(ctx).Where(app.QueueSignal{ID: qsID}).First(&signal).Error; err == nil {
			asi.Signal = &signal
		}

		awaited = append(awaited, asi)
	}

	return awaited
}

// ExtractEnqueuedSignals looks for EnqueueSignal activities and extracts the created signal IDs.
func (h *Helpers) ExtractEnqueuedSignals(ctx context.Context, activities []ActivityInfo) []EnqueuedSignalInfo {
	var enqueued []EnqueuedSignalInfo

	for _, act := range activities {
		if !strings.Contains(act.Name, "EnqueueSignal") {
			continue
		}

		qsID := ExtractQueueSignalIDFromResult(act.Result)
		if qsID == "" {
			qsID = ExtractQueueSignalIDFromInput(act.Input)
		}
		if qsID == "" {
			continue
		}

		esi := EnqueuedSignalInfo{
			QueueSignalID: qsID,
			ActivityName:  act.Name,
		}

		var signal app.QueueSignal
		if err := h.db.WithContext(ctx).Where(app.QueueSignal{ID: qsID}).First(&signal).Error; err == nil {
			esi.Signal = &signal
		}

		enqueued = append(enqueued, esi)
	}

	return enqueued
}

// ExtractQueueSignalIDFromResult parses activity result JSON for a queue signal ID.
func ExtractQueueSignalIDFromResult(result string) string {
	result = strings.TrimSpace(result)
	if result == "" {
		return ""
	}

	var id string
	if err := json.Unmarshal([]byte(result), &id); err == nil {
		if strings.HasPrefix(id, "qsi") {
			return id
		}
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err == nil {
		for _, key := range []string{"id", "ID", "queue_signal_id", "QueueSignalID"} {
			if id, ok := obj[key].(string); ok && strings.HasPrefix(id, "qsi") {
				return id
			}
		}
	}

	return ""
}

// ExtractQueueSignalIDFromInput parses activity input JSON for a queue signal ID.
func ExtractQueueSignalIDFromInput(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var id string
	if err := json.Unmarshal([]byte(input), &id); err == nil {
		if strings.HasPrefix(id, "qsi") {
			return id
		}
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(input), &obj); err == nil {
		for _, key := range []string{"queue_signal_id", "QueueSignalID", "id", "ID"} {
			if id, ok := obj[key].(string); ok && strings.HasPrefix(id, "qsi") {
				return id
			}
		}
	}

	return ""
}

// FormatWorkflowStatus converts a Temporal workflow execution status to a human-readable string.
func FormatWorkflowStatus(status enumspb.WorkflowExecutionStatus) string {
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

func buildActivityInfo(
	scheduled map[int64]scheduledActivity,
	started map[int64]time.Time,
	attempts map[int64]int32,
	scheduledEventID int64,
	event *historypb.HistoryEvent,
	status string,
	failure string,
) ActivityInfo {
	ai := ActivityInfo{
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

// BuildUpdateExecutions groups activities into their parent update executions.
func BuildUpdateExecutions(updates map[int64]*updateState, activities []ActivityInfo, awaitedSignals []AwaitedSignalInfo, enqueuedSignals []EnqueuedSignalInfo) ([]UpdateExecution, []ActivityInfo) {
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

	actsByUpdate := make(map[int64][]ActivityInfo)
	var orphans []ActivityInfo

	for _, act := range activities {
		ownerIdx := -1
		for i, su := range sorted {
			if act.ScheduledEventID <= su.acceptedEvtID {
				break
			}
			if i+1 < len(sorted) && act.ScheduledEventID >= sorted[i+1].acceptedEvtID {
				continue
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

	awaitedByID := make(map[string]AwaitedSignalInfo)
	for _, as := range awaitedSignals {
		if as.QueueSignalID != "" {
			awaitedByID[as.QueueSignalID] = as
		}
	}
	enqueuedByID := make(map[string]EnqueuedSignalInfo)
	for _, es := range enqueuedSignals {
		if es.QueueSignalID != "" {
			enqueuedByID[es.QueueSignalID] = es
		}
	}

	execs := make([]UpdateExecution, 0, len(sorted))
	for _, su := range sorted {
		ue := UpdateExecution{
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

		for _, act := range ue.Activities {
			if strings.Contains(act.Name, "AwaitSignal") {
				qsID := ExtractQueueSignalIDFromInput(act.Input)
				if as, ok := awaitedByID[qsID]; ok {
					ue.AwaitedSignals = append(ue.AwaitedSignals, as)
				}
			}
			if strings.Contains(act.Name, "EnqueueSignal") {
				qsID := ExtractQueueSignalIDFromResult(act.Result)
				if qsID == "" {
					qsID = ExtractQueueSignalIDFromInput(act.Input)
				}
				if es, ok := enqueuedByID[qsID]; ok {
					ue.EnqueuedSignals = append(ue.EnqueuedSignals, es)
				}
			}
		}

		execs = append(execs, ue)
	}

	return execs, orphans
}

// formatPayloadsSimple extracts raw JSON from Temporal payloads without codec decoding.
// This is a simplified version that works without codec dependencies.
func formatPayloadsSimple(payloads *commonpb.Payloads) string {
	if payloads == nil {
		return ""
	}
	for _, p := range payloads.GetPayloads() {
		if p == nil || len(p.GetData()) == 0 {
			continue
		}
		if json.Valid(p.GetData()) {
			return string(p.GetData())
		}
	}
	return ""
}
