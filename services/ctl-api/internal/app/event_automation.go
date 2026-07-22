package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type EventSourceAuthType string
type EventEnvelopeType string
type EventSourceStatus string
type EventRoutingStatus string
type EventAutomationFilterType string
type EventAutomationTargetType string
type EventDispatchStatus string

const (
	EventSourceAuthTypeNone         EventSourceAuthType = "none"
	EventSourceAuthTypeHMAC         EventSourceAuthType = "hmac"
	EventSourceAuthTypeAPIKey       EventSourceAuthType = "api_key"
	EventSourceAuthTypeBasic        EventSourceAuthType = "basic"
	EventSourceAuthTypeBearerJWT    EventSourceAuthType = "bearer_jwt"
	EventSourceAuthTypeSNSSignature EventSourceAuthType = "sns_signature"

	EventEnvelopeTypeNone        EventEnvelopeType = "none"
	EventEnvelopeTypePubSubPush  EventEnvelopeType = "pubsub_push"
	EventEnvelopeTypeCloudEvents EventEnvelopeType = "cloudevents"
	EventEnvelopeTypeSNS         EventEnvelopeType = "sns"

	EventSourceStatusActive    EventSourceStatus = "active"
	EventSourceStatusSuspended EventSourceStatus = "suspended"

	EventRoutingStatusAccepted      EventRoutingStatus = "accepted"
	EventRoutingStatusRouting       EventRoutingStatus = "routing"
	EventRoutingStatusMatched       EventRoutingStatus = "matched"
	EventRoutingStatusIgnored       EventRoutingStatus = "ignored"
	EventRoutingStatusRejected      EventRoutingStatus = "rejected"
	EventRoutingStatusRoutingFailed EventRoutingStatus = "routing_failed"

	EventAutomationFilterTypeEq        EventAutomationFilterType = "eq"
	EventAutomationFilterTypeNEq       EventAutomationFilterType = "neq"
	EventAutomationFilterTypeIn        EventAutomationFilterType = "in"
	EventAutomationFilterTypePrefix    EventAutomationFilterType = "prefix"
	EventAutomationFilterTypeSuffix    EventAutomationFilterType = "suffix"
	EventAutomationFilterTypeContains  EventAutomationFilterType = "contains"
	EventAutomationFilterTypeGT        EventAutomationFilterType = "gt"
	EventAutomationFilterTypeGTE       EventAutomationFilterType = "gte"
	EventAutomationFilterTypeLT        EventAutomationFilterType = "lt"
	EventAutomationFilterTypeLTE       EventAutomationFilterType = "lte"
	EventAutomationFilterTypeRegex     EventAutomationFilterType = "regex"
	EventAutomationFilterTypeExists    EventAutomationFilterType = "exists"
	EventAutomationFilterTypeNotExists EventAutomationFilterType = "not_exists"

	EventAutomationTargetTypeAppBranchRun EventAutomationTargetType = "app_branch_run"

	EventDispatchStatusPending         EventDispatchStatus = "pending"
	EventDispatchStatusDispatching     EventDispatchStatus = "dispatching"
	EventDispatchStatusTriggered       EventDispatchStatus = "triggered"
	EventDispatchStatusRetryableFailed EventDispatchStatus = "retryable_failed"
	EventDispatchStatusDeadLettered    EventDispatchStatus = "dead_lettered"
	EventDispatchStatusCancelled       EventDispatchStatus = "cancelled"
)

type EventAutomationFilter struct {
	From  string                    `json:"from,omitempty"`
	Op    EventAutomationFilterType `json:"op"`
	Path  string                    `json:"path"`
	Value any                       `json:"value"`
}

type EventFieldSelector struct {
	Header  string `json:"header,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type EventSourceAuthConfig struct {
	Header    string   `json:"header,omitempty"`
	Prefix    string   `json:"prefix,omitempty"`
	Encoding  string   `json:"encoding,omitempty"`
	Algorithm string   `json:"algorithm,omitempty"`
	Username  string   `json:"username,omitempty"`
	Issuer    string   `json:"issuer,omitempty"`
	Audience  []string `json:"audience,omitempty"`
	TopicARN  string   `json:"topic_arn,omitempty"`
}

func (e *EventAutomationFilter) UnmarshalJSON(data []byte) error {
	var encoded struct {
		From  string                    `json:"from"`
		Op    EventAutomationFilterType `json:"op"`
		Path  string                    `json:"path"`
		Value json.RawMessage           `json:"value"`
	}
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	var value any
	if len(encoded.Value) != 0 {
		decoder := json.NewDecoder(bytes.NewReader(encoded.Value))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	e.Op = encoded.Op
	e.From = encoded.From
	e.Path = encoded.Path
	e.Value = value
	return nil
}

type EventSource struct {
	ID             string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID    string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt      time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt      time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt      soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID          string                `json:"org_id,omitzero" gorm:"notnull;<-:create" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org            Org                   `json:"-" temporaljson:"-"`
	IngressKeyHash string                `json:"-" gorm:"notnull;<-:create" temporaljson:"-"`
	Name           string                `json:"name" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Description    string                `json:"description,omitempty" temporaljson:"description,omitzero,omitempty"`
	AuthType       EventSourceAuthType   `json:"auth_type" gorm:"notnull;<-:create;check:event_source_auth_type_checker,auth_type IN ('none','hmac','api_key','basic','bearer_jwt','sns_signature')" temporaljson:"auth_type,omitzero,omitempty"`
	AuthConfig     EventSourceAuthConfig `json:"auth_config,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"auth_config,omitzero,omitempty"`
	Envelope       EventEnvelopeType     `json:"envelope" gorm:"notnull;<-:create;check:event_source_envelope_checker,envelope IN ('none','pubsub_push','cloudevents','sns')" temporaljson:"envelope,omitzero,omitempty"`
	TypeFrom       EventFieldSelector    `json:"type_from,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"type_from,omitzero,omitempty"`
	IDFrom         EventFieldSelector    `json:"id_from,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"id_from,omitzero,omitempty"`
	Status         EventSourceStatus     `json:"status" gorm:"notnull;check:event_source_status_checker,status IN ('active','suspended')" temporaljson:"status,omitzero,omitempty"`
	LastEventAt    *time.Time            `json:"last_event_at,omitempty" temporaljson:"last_event_at,omitzero,omitempty"`
	Secrets        []EventSourceSecret   `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	Events         []EventSourceEvent    `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
}

func (e *EventSource) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "org_id_name_deleted_at"), Columns: []string{"org_id", "name", "deleted_at"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "ingress_key_hash"), Columns: []string{"ingress_key_hash"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
	}
}

func (e *EventSource) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewEventSourceID()
	}
	if e.OrgID == "" {
		e.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if e.CreatedByID == "" {
		e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if e.AuthType == "" {
		e.AuthType = EventSourceAuthTypeHMAC
	}
	if e.Envelope == "" {
		e.Envelope = EventEnvelopeTypeNone
	}
	if e.Status == "" {
		e.Status = EventSourceStatusActive
	}
	return nil
}

type EventSourceSecret struct {
	ID            string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID   string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt     time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt     time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt     soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID         string                `json:"-" gorm:"notnull;<-:create" temporaljson:"-"`
	Org           Org                   `json:"-" temporaljson:"-"`
	EventSourceID string                `json:"event_source_id" gorm:"notnull;<-:create" temporaljson:"event_source_id,omitzero,omitempty"`
	EventSource   EventSource           `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	KeyID         string                `json:"key_id" gorm:"notnull;<-:create" temporaljson:"key_id,omitzero,omitempty"`
	Secret        string                `json:"-" gorm:"notnull;<-:create" temporaljson:"-"`
	NotBefore     time.Time             `json:"not_before" gorm:"notnull;<-:create" temporaljson:"not_before,omitzero,omitempty"`
	ExpiresAt     *time.Time            `json:"expires_at,omitempty" gorm:"check:event_source_secret_expiration_checker,expires_at IS NULL OR expires_at > not_before" temporaljson:"expires_at,omitzero,omitempty"`
	RevokedAt     *time.Time            `json:"revoked_at,omitempty" temporaljson:"revoked_at,omitzero,omitempty"`
	LastUsedAt    *time.Time            `json:"last_used_at,omitempty" temporaljson:"last_used_at,omitzero,omitempty"`
}

func (e *EventSourceSecret) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "event_source_id_key_id"), Columns: []string{"event_source_id", "key_id"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "event_source_id"), Columns: []string{"event_source_id"}},
	}
}

func (e *EventSourceSecret) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewEventSourceSecretID()
	}
	if e.KeyID == "" {
		e.KeyID = e.ID
	}
	if e.CreatedByID == "" {
		e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if e.OrgID == "" {
		e.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type EventSourceEvent struct {
	ID                  string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt           time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt           time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt           soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	EventSourceID       string                `json:"event_source_id" gorm:"notnull;<-:create" temporaljson:"event_source_id,omitzero,omitempty"`
	EventSource         EventSource           `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	EventSourceSecretID *string               `json:"event_source_secret_id,omitempty" gorm:"<-:create" temporaljson:"event_source_secret_id,omitzero,omitempty"`
	EventSourceSecret   EventSourceSecret     `json:"-" temporaljson:"-"`
	OrgID               string                `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org                 Org                   `json:"-" temporaljson:"-"`
	ExternalID          string                `json:"external_id" gorm:"notnull;<-:create" temporaljson:"external_id,omitzero,omitempty"`
	EventType           string                `json:"event_type" gorm:"notnull;<-:create" temporaljson:"event_type,omitzero,omitempty"`
	OccurredAt          *time.Time            `json:"occurred_at,omitempty" gorm:"<-:create" temporaljson:"occurred_at,omitzero,omitempty"`
	ReceivedAt          time.Time             `json:"received_at" gorm:"notnull;<-:create" temporaljson:"received_at,omitzero,omitempty"`
	Payload             json.RawMessage       `json:"payload" gorm:"serializer:json;type:jsonb;notnull;<-:create" temporaljson:"payload,omitzero,omitempty"`
	Headers             map[string][]string   `json:"headers,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"headers,omitzero,omitempty"`
	RawBody             []byte                `json:"-" gorm:"type:bytea;notnull;<-:create" temporaljson:"-"`
	RawBodySHA256       string                `json:"raw_body_sha256" gorm:"notnull;<-:create" temporaljson:"raw_body_sha256,omitzero,omitempty"`
	RawBodySize         int64                 `json:"raw_body_size" gorm:"<-:create" temporaljson:"raw_body_size,omitzero,omitempty"`
	RawContentType      string                `json:"raw_content_type,omitempty" gorm:"<-:create" temporaljson:"raw_content_type,omitzero,omitempty"`
	PayloadContentType  string                `json:"payload_content_type,omitempty" gorm:"<-:create" temporaljson:"payload_content_type,omitzero,omitempty"`
	SecretKeyID         string                `json:"secret_key_id,omitempty" gorm:"<-:create" temporaljson:"secret_key_id,omitzero,omitempty"`
	RoutingStatus       EventRoutingStatus    `json:"routing_status" gorm:"notnull;check:event_routing_status_checker,routing_status IN ('accepted','routing','matched','ignored','rejected','routing_failed')" temporaljson:"routing_status,omitzero,omitempty"`
	RoutingError        string                `json:"routing_error,omitempty" temporaljson:"routing_error,omitzero,omitempty"`
	RoutingStartedAt    *time.Time            `json:"routing_started_at,omitempty" temporaljson:"routing_started_at,omitzero,omitempty"`
	RoutingCompletedAt  *time.Time            `json:"routing_completed_at,omitempty" temporaljson:"routing_completed_at,omitzero,omitempty"`
	MatchCount          int                   `json:"match_count" temporaljson:"match_count,omitzero,omitempty"`
	DispatchCount       int                   `json:"dispatch_count" temporaljson:"dispatch_count,omitzero,omitempty"`
}

func (e *EventSourceEvent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "event_source_id_external_id"), Columns: []string{"event_source_id", "external_id"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "org_id_received_at"), Columns: []string{"org_id", "received_at"}},
		{Name: indexes.Name(db, e, "routing_status_received_at"), Columns: []string{"routing_status", "received_at"}},
	}
}
func (e *EventSourceEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewEventSourceEventID()
	}
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now()
	}
	if e.RoutingStatus == "" {
		e.RoutingStatus = EventRoutingStatusAccepted
	}
	return nil
}

type EventAutomationRule struct {
	ID            string                    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID   string                    `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt     time.Time                 `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt     time.Time                 `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt     soft_delete.DeletedAt     `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID         string                    `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org           Org                       `json:"-" temporaljson:"-"`
	AppID         string                    `json:"app_id" gorm:"notnull;<-:create" temporaljson:"app_id,omitzero,omitempty"`
	App           App                       `json:"-" temporaljson:"-"`
	AppConfigID   string                    `json:"app_config_id" gorm:"notnull;<-:create" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig     AppConfig                 `json:"-" temporaljson:"app_config,omitzero,omitempty"`
	EventSourceID string                    `json:"event_source_id" gorm:"notnull;<-:create" temporaljson:"event_source_id,omitzero,omitempty"`
	EventSource   EventSource               `json:"-" temporaljson:"-"`
	Name          string                    `json:"name" gorm:"notnull;<-:create" temporaljson:"name,omitzero,omitempty"`
	Enabled       bool                      `json:"enabled" gorm:"notnull;<-:create" temporaljson:"enabled,omitzero,omitempty"`
	SuspendedAt   *time.Time                `json:"suspended_at,omitempty" temporaljson:"suspended_at,omitzero,omitempty"`
	SuspendedByID *string                   `json:"suspended_by_id,omitempty" temporaljson:"suspended_by_id,omitzero,omitempty"`
	ValidFrom     time.Time                 `json:"valid_from" gorm:"notnull;<-:create" temporaljson:"valid_from,omitzero,omitempty"`
	ValidTo       *time.Time                `json:"valid_to,omitempty" gorm:"check:event_automation_rule_validity_checker,valid_to IS NULL OR valid_to > valid_from" temporaljson:"valid_to,omitzero,omitempty"`
	EventTypes    pq.StringArray            `json:"event_types,omitempty" gorm:"type:text[];<-:create" temporaljson:"event_types,omitzero,omitempty"`
	Filters       []EventAutomationFilter   `json:"filters" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"filters,omitzero,omitempty"`
	TargetType    EventAutomationTargetType `json:"target_type" gorm:"notnull;<-:create;check:event_automation_rule_target_type_checker,target_type IN ('app_branch_run')" temporaljson:"target_type,omitzero,omitempty"`
	AppBranchID   string                    `json:"app_branch_id" gorm:"notnull;<-:create" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch     AppBranch                 `json:"-" temporaljson:"-"`
	Force         bool                      `json:"force" gorm:"<-:create" temporaljson:"force,omitzero,omitempty"`
	PlanOnly      bool                      `json:"plan_only" gorm:"<-:create" temporaljson:"plan_only,omitzero,omitempty"`
	ConfigHash    string                    `json:"config_hash" gorm:"notnull;<-:create" temporaljson:"config_hash,omitzero,omitempty"`
}

func (e *EventAutomationRule) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "app_config_id_name_deleted_at"), Columns: []string{"app_config_id", "name", "deleted_at"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "event_source_id_valid_from"), Columns: []string{"event_source_id", "valid_from"}},
	}
}

func (e *EventAutomationRule) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewEventAutomationRuleID()
	}
	if e.OrgID == "" {
		e.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if e.CreatedByID == "" {
		e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

type EventDispatch struct {
	ID                    string                    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID           string                    `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt             time.Time                 `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt             time.Time                 `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	OrgID                 string                    `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org                   Org                       `json:"-" temporaljson:"-"`
	AppID                 string                    `json:"app_id" gorm:"notnull;<-:create" temporaljson:"app_id,omitzero,omitempty"`
	App                   App                       `json:"-" temporaljson:"-"`
	EventSourceEventID    string                    `json:"event_source_event_id" gorm:"notnull;<-:create" temporaljson:"event_source_event_id,omitzero,omitempty"`
	EventSourceEvent      EventSourceEvent          `json:"-" temporaljson:"-"`
	EventAutomationRuleID string                    `json:"event_automation_rule_id" gorm:"notnull;<-:create" temporaljson:"event_automation_rule_id,omitzero,omitempty"`
	EventAutomationRule   EventAutomationRule       `json:"-" temporaljson:"-"`
	ReplayID              *string                   `json:"replay_id,omitempty" gorm:"<-:create" temporaljson:"replay_id,omitzero,omitempty"`
	IdempotencyKey        string                    `json:"idempotency_key" gorm:"notnull;<-:create" temporaljson:"idempotency_key,omitzero,omitempty"`
	TargetType            EventAutomationTargetType `json:"target_type" gorm:"notnull;<-:create;check:event_dispatch_target_type_checker,target_type IN ('app_branch_run')" temporaljson:"target_type,omitzero,omitempty"`
	TargetID              string                    `json:"target_id" gorm:"notnull;<-:create" temporaljson:"target_id,omitzero,omitempty"`
	MappedInputs          map[string]any            `json:"mapped_inputs,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"mapped_inputs,omitzero,omitempty"`
	Status                EventDispatchStatus       `json:"status" gorm:"notnull;check:event_dispatch_status_checker,status IN ('pending','dispatching','triggered','retryable_failed','dead_lettered','cancelled')" temporaljson:"status,omitzero,omitempty"`
	Attempts              int                       `json:"attempts" gorm:"check:event_dispatch_attempts_checker,attempts >= 0" temporaljson:"attempts,omitzero,omitempty"`
	GenerationToken       string                    `json:"-" temporaljson:"-"`
	ExecutionToken        string                    `json:"-" temporaljson:"-"`
	NextAttemptAt         *time.Time                `json:"next_attempt_at,omitempty" temporaljson:"next_attempt_at,omitzero,omitempty"`
	Error                 string                    `json:"error,omitempty" temporaljson:"error,omitzero,omitempty"`
	QueueSignalID         *string                   `json:"queue_signal_id,omitempty" temporaljson:"queue_signal_id,omitzero,omitempty"`
	ResultResourceType    string                    `json:"result_resource_type,omitempty" temporaljson:"result_resource_type,omitzero,omitempty"`
	ResultResourceID      string                    `json:"result_resource_id,omitempty" temporaljson:"result_resource_id,omitzero,omitempty"`
	WorkflowID            string                    `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
	StartedAt             *time.Time                `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`
	TriggeredAt           *time.Time                `json:"triggered_at,omitempty" temporaljson:"triggered_at,omitzero,omitempty"`
	FailedAt              *time.Time                `json:"failed_at,omitempty" temporaljson:"failed_at,omitzero,omitempty"`
}

func (e *EventDispatch) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "idempotency_key"), Columns: []string{"idempotency_key"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "event_source_event_id"), Columns: []string{"event_source_event_id"}},
		{Name: indexes.Name(db, e, "status_next_attempt_at"), Columns: []string{"status", "next_attempt_at"}},
	}
}
func (e *EventDispatch) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewEventDispatchID()
	}
	if e.OrgID == "" {
		e.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if e.CreatedByID == "" {
		e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if e.Status == "" {
		e.Status = EventDispatchStatusPending
	}
	if e.GenerationToken == "" {
		e.GenerationToken = uuid.NewString()
	}
	return nil
}
