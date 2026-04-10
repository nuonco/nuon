package service

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	}

	component := views.QueueSignalDetail(&signal, &q, s.cfg.TemporalUIURL, wfInfo)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

type scheduledActivity struct {
	name        string
	scheduledAt time.Time
	input       string // JSON-formatted activity input
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

	var activities []views.ActivityInfo

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			s.l.Warn("failed to iterate workflow history", zap.Error(err))
			break
		}

		switch event.GetEventType() {
		case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
			attrs := event.GetActivityTaskScheduledEventAttributes()
			if attrs != nil {
				name := ""
				if attrs.GetActivityType() != nil {
					name = attrs.GetActivityType().GetName()
				}
				scheduled[event.GetEventId()] = scheduledActivity{
					name:        name,
					scheduledAt: event.GetEventTime().AsTime(),
					input:       s.formatPayloads(attrs.GetInput()),
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
		}
	}

	// Add any scheduled-but-not-finished activities as "Running" or "Scheduled"
	for schedID, sched := range scheduled {
		found := false
		for _, a := range activities {
			if a.Name == sched.name {
				found = true
				break
			}
		}
		if !found {
			actInfo := views.ActivityInfo{
				Name: sched.name,
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
	return info
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
		Status:     status,
		FinishedAt: event.GetEventTime().AsTime(),
		Failure:    failure,
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
