package models

import (
	"encoding/json"
	"time"
)

type TriggerEvent struct {
	ID                    string                       `json:"id"`
	CreatedAt             time.Time                    `json:"created_at"`
	UpdatedAt             time.Time                    `json:"updated_at"`
	TriggerID             string                       `json:"trigger_id"`
	TriggerName           string                       `json:"trigger_name,omitempty"`
	TriggerSecretID       *string                      `json:"trigger_secret_id,omitempty"`
	OrgID                 string                       `json:"org_id"`
	ExternalID            string                       `json:"external_id"`
	Source                string                       `json:"source,omitempty"`
	EventType             string                       `json:"event_type"`
	OccurredAt            *time.Time                   `json:"occurred_at,omitempty"`
	ReceivedAt            time.Time                    `json:"received_at"`
	Payload               json.RawMessage              `json:"payload"`
	Headers               map[string][]string          `json:"headers,omitempty"`
	RawBodySHA256         string                       `json:"raw_body_sha256"`
	PayloadSHA256         string                       `json:"payload_sha256"`
	RawBodySize           int64                        `json:"raw_body_size"`
	RawContentType        string                       `json:"raw_content_type,omitempty"`
	PayloadContentType    string                       `json:"payload_content_type,omitempty"`
	SecretKeyID           string                       `json:"secret_key_id,omitempty"`
	RoutingStatus         string                       `json:"routing_status"`
	RoutingError          string                       `json:"routing_error,omitempty"`
	RoutingStartedAt      *time.Time                   `json:"routing_started_at,omitempty"`
	RoutingCompletedAt    *time.Time                   `json:"routing_completed_at,omitempty"`
	MatchCount            int                          `json:"match_count"`
	WaiterMatchCount      int                          `json:"waiter_match_count"`
	DispatchCount         int                          `json:"dispatch_count"`
	MatchExplanations     []TriggerEventRuleEvaluation `json:"match_explanations,omitempty"`
	ExplanationsTruncated bool                         `json:"explanations_truncated,omitempty"`
	Dispatches            []TriggerEventDispatch       `json:"dispatches,omitempty"`
	DispatchesTruncated   bool                         `json:"dispatches_truncated,omitempty"`
	WaiterMatches         []TriggerEventWaiterMatch    `json:"waiter_matches,omitempty"`
}

type TriggerEventWaiterMatch struct {
	ID               string          `json:"id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	OrgID            string          `json:"org_id"`
	AppID            string          `json:"app_id"`
	InstallID        string          `json:"install_id"`
	WorkflowID       string          `json:"workflow_id"`
	WorkflowStepID   string          `json:"workflow_step_id"`
	QueueSignalID    string          `json:"queue_signal_id"`
	TriggerID        string          `json:"trigger_id"`
	EventTypes       []string        `json:"event_types"`
	Filters          json.RawMessage `json:"filters,omitempty"`
	Status           string          `json:"status"`
	MatchedEventID   *string         `json:"matched_event_id,omitempty"`
	ActivatedAt      time.Time       `json:"activated_at"`
	MatchedAt        *time.Time      `json:"matched_at,omitempty"`
	NotifiedAt       *time.Time      `json:"notified_at,omitempty"`
	CancelledAt      *time.Time      `json:"cancelled_at,omitempty"`
	ExpiredAt        *time.Time      `json:"expired_at,omitempty"`
	TriggerName      string          `json:"trigger_name,omitempty"`
	MatchedEventType string          `json:"matched_event_type,omitempty"`
	WorkflowStepName string          `json:"workflow_step_name,omitempty"`
	RunbookRunID     string          `json:"runbook_run_id,omitempty"`
	RunbookID        string          `json:"runbook_id,omitempty"`
	RunbookName      string          `json:"runbook_name,omitempty"`
}

type TriggerEventSummary struct {
	ID                  string                        `json:"id"`
	TriggerID           string                        `json:"trigger_id"`
	TriggerName         string                        `json:"trigger_name,omitempty"`
	ExternalID          string                        `json:"external_id"`
	Source              string                        `json:"source,omitempty"`
	EventType           string                        `json:"event_type"`
	OccurredAt          *time.Time                    `json:"occurred_at,omitempty"`
	ReceivedAt          time.Time                     `json:"received_at"`
	RoutingStatus       string                        `json:"routing_status"`
	RoutingError        string                        `json:"routing_error,omitempty"`
	RoutingStartedAt    *time.Time                    `json:"routing_started_at,omitempty"`
	RoutingCompletedAt  *time.Time                    `json:"routing_completed_at,omitempty"`
	MatchCount          int                           `json:"match_count"`
	WaiterMatchCount    int                           `json:"waiter_match_count"`
	DispatchCount       int                           `json:"dispatch_count"`
	Dispatches          []TriggerEventDispatchSummary `json:"dispatches"`
	DispatchesTruncated bool                          `json:"dispatches_truncated"`
}

type TriggerEventPage struct {
	Items      []*TriggerEventSummary `json:"items"`
	NextCursor string                 `json:"next_cursor,omitempty"`
}

type TriggerEventListQuery struct {
	Limit          int
	Trigger        string
	EventType      string
	Outcome        string
	Search         string
	ReceivedAfter  string
	ReceivedBefore string
	Cursor         string
}

type TriggerEventDispatchSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type TriggerEventRaw struct {
	RawBodyBase64  string `json:"raw_body_base64"`
	RawBodySHA256  string `json:"raw_body_sha256"`
	RawBodySize    int64  `json:"raw_body_size"`
	RawContentType string `json:"raw_content_type,omitempty"`
}

type TriggerEventRuleEvaluation struct {
	RuleID            string                         `json:"rule_id"`
	RuleName          string                         `json:"rule_name"`
	AppID             string                         `json:"app_id"`
	EventType         string                         `json:"event_type,omitempty"`
	AllowedEventTypes []string                       `json:"allowed_event_types,omitempty"`
	EventTypeMatched  bool                           `json:"event_type_matched"`
	Filters           []TriggerEventFilterEvaluation `json:"filters,omitempty"`
	Matched           bool                           `json:"matched"`
}

type TriggerEventFilterEvaluation struct {
	From      string `json:"from"`
	Path      string `json:"path"`
	Op        string `json:"op"`
	Expected  any    `json:"expected,omitempty"`
	Selected  []any  `json:"selected,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Matched   bool   `json:"matched"`
	Error     string `json:"error,omitempty"`
}

type TriggerEventDispatch struct {
	ID                 string            `json:"id"`
	CreatedByID        string            `json:"created_by_id"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	OrgID              string            `json:"org_id"`
	AppID              string            `json:"app_id"`
	TriggerEventID     string            `json:"trigger_event_id"`
	TriggerRuleID      string            `json:"trigger_rule_id"`
	ReplayID           *string           `json:"replay_id,omitempty"`
	IdempotencyKey     string            `json:"idempotency_key"`
	TargetType         string            `json:"target_type"`
	TargetID           string            `json:"target_id"`
	RunbookConfigID    *string           `json:"runbook_config_id,omitempty"`
	MappedInputs       map[string]string `json:"mapped_inputs,omitempty"`
	Status             string            `json:"status"`
	Attempts           int               `json:"attempts"`
	NextAttemptAt      *time.Time        `json:"next_attempt_at,omitempty"`
	Error              string            `json:"error,omitempty"`
	QueueSignalID      *string           `json:"queue_signal_id,omitempty"`
	ResultResourceType string            `json:"result_resource_type,omitempty"`
	ResultResourceID   string            `json:"result_resource_id,omitempty"`
	WorkflowID         string            `json:"workflow_id,omitempty"`
	StartedAt          *time.Time        `json:"started_at,omitempty"`
	TriggeredAt        *time.Time        `json:"triggered_at,omitempty"`
	FailedAt           *time.Time        `json:"failed_at,omitempty"`
	InstallID          string            `json:"install_id,omitempty"`
	RunbookID          string            `json:"runbook_id,omitempty"`
	RunbookName        string            `json:"runbook_name,omitempty"`
}

type TriggerEventDispatchPage struct {
	Items      []*TriggerEventDispatch `json:"items"`
	NextCursor string                  `json:"next_cursor,omitempty"`
}

type TriggerEventReplayResponse struct {
	EventID  string `json:"event_id"`
	ReplayID string `json:"replay_id"`
}

type TriggerEventDispatchRetryResponse struct {
	DispatchID string `json:"dispatch_id"`
	RetryID    string `json:"retry_id"`
}
