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

type TriggerAuthType string
type EventEnvelopeType string
type TriggerStatus string
type EventRoutingStatus string
type TriggerFilterType string
type TriggerTargetType string
type EventDispatchStatus string

const (
	TriggerAuthTypeNone         TriggerAuthType = "none"
	TriggerAuthTypeHMAC         TriggerAuthType = "hmac"
	TriggerAuthTypeAPIKey       TriggerAuthType = "api_key"
	TriggerAuthTypeBasic        TriggerAuthType = "basic"
	TriggerAuthTypeBearerJWT    TriggerAuthType = "bearer_jwt"
	TriggerAuthTypeSNSSignature TriggerAuthType = "sns_signature"

	EventEnvelopeTypeNone        EventEnvelopeType = "none"
	EventEnvelopeTypePubSubPush  EventEnvelopeType = "pubsub_push"
	EventEnvelopeTypeCloudEvents EventEnvelopeType = "cloudevents"
	EventEnvelopeTypeSNS         EventEnvelopeType = "sns"

	TriggerStatusActive    TriggerStatus = "active"
	TriggerStatusSuspended TriggerStatus = "suspended"

	EventRoutingStatusAccepted      EventRoutingStatus = "accepted"
	EventRoutingStatusRouting       EventRoutingStatus = "routing"
	EventRoutingStatusMatched       EventRoutingStatus = "matched"
	EventRoutingStatusIgnored       EventRoutingStatus = "ignored"
	EventRoutingStatusRejected      EventRoutingStatus = "rejected"
	EventRoutingStatusRoutingFailed EventRoutingStatus = "routing_failed"

	TriggerFilterTypeEq        TriggerFilterType = "eq"
	TriggerFilterTypeNEq       TriggerFilterType = "neq"
	TriggerFilterTypeIn        TriggerFilterType = "in"
	TriggerFilterTypePrefix    TriggerFilterType = "prefix"
	TriggerFilterTypeSuffix    TriggerFilterType = "suffix"
	TriggerFilterTypeContains  TriggerFilterType = "contains"
	TriggerFilterTypeGT        TriggerFilterType = "gt"
	TriggerFilterTypeGTE       TriggerFilterType = "gte"
	TriggerFilterTypeLT        TriggerFilterType = "lt"
	TriggerFilterTypeLTE       TriggerFilterType = "lte"
	TriggerFilterTypeRegex     TriggerFilterType = "regex"
	TriggerFilterTypeExists    TriggerFilterType = "exists"
	TriggerFilterTypeNotExists TriggerFilterType = "not_exists"

	TriggerTargetTypeAppBranchRun TriggerTargetType = "app_branch_run"
	TriggerTargetTypeRunbook      TriggerTargetType = "runbook"

	EventDispatchStatusPending         EventDispatchStatus = "pending"
	EventDispatchStatusDispatching     EventDispatchStatus = "dispatching"
	EventDispatchStatusTriggered       EventDispatchStatus = "triggered"
	EventDispatchStatusRetryableFailed EventDispatchStatus = "retryable_failed"
	EventDispatchStatusDeadLettered    EventDispatchStatus = "dead_lettered"
	EventDispatchStatusCancelled       EventDispatchStatus = "cancelled"
)

type TriggerFilter struct {
	From  string            `json:"from,omitempty"`
	Op    TriggerFilterType `json:"op"`
	Path  string            `json:"path"`
	Value any               `json:"value"`
}

type TriggerFilterEvaluation struct {
	From      string            `json:"from"`
	Path      string            `json:"path"`
	Op        TriggerFilterType `json:"op"`
	Expected  any               `json:"expected,omitempty"`
	Selected  []any             `json:"selected,omitempty"`
	Truncated bool              `json:"truncated,omitempty"`
	Matched   bool              `json:"matched"`
	Error     string            `json:"error,omitempty"`
}

type TriggerRuleEvaluation struct {
	RuleID            string                    `json:"rule_id"`
	RuleName          string                    `json:"rule_name"`
	AppID             string                    `json:"app_id"`
	EventType         string                    `json:"event_type,omitempty"`
	AllowedEventTypes []string                  `json:"allowed_event_types,omitempty"`
	EventTypeMatched  bool                      `json:"event_type_matched"`
	Filters           []TriggerFilterEvaluation `json:"filters,omitempty"`
	Matched           bool                      `json:"matched"`
}

type EventFieldSelector struct {
	Header  string `json:"header,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type TriggerAuthConfig struct {
	Header          string   `json:"header,omitempty"`
	Prefix          string   `json:"prefix,omitempty"`
	Encoding        string   `json:"encoding,omitempty"`
	Algorithm       string   `json:"algorithm,omitempty"`
	Username        string   `json:"username,omitempty"`
	Issuer          string   `json:"issuer,omitempty"`
	Audience        []string `json:"audience,omitempty"`
	TopicARN        string   `json:"topic_arn,omitempty"`
	ExpectedEmail   string   `json:"expected_email,omitempty"`
	ExpectedSubject string   `json:"expected_subject,omitempty"`
}

func (e *TriggerFilter) UnmarshalJSON(data []byte) error {
	var encoded struct {
		From  string            `json:"from"`
		Op    TriggerFilterType `json:"op"`
		Path  string            `json:"path"`
		Value json.RawMessage   `json:"value"`
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

type Trigger struct {
	ID             string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID    string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt      time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt      time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt      soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID          string                `json:"org_id,omitzero" gorm:"notnull;<-:create" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org            Org                   `json:"-" temporaljson:"-"`
	IngressKey     string                `json:"-" temporaljson:"-"`
	IngressKeyHash string                `json:"-" gorm:"notnull" temporaljson:"-"`
	Name           string                `json:"name" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Description    string                `json:"description,omitempty" temporaljson:"description,omitzero,omitempty"`
	Preset         string                `json:"preset,omitempty" gorm:"<-:create" temporaljson:"preset,omitzero,omitempty"`
	AuthType       TriggerAuthType       `json:"auth_type" gorm:"notnull;<-:create;check:trigger_auth_type_checker,auth_type IN ('none','hmac','api_key','basic','bearer_jwt','sns_signature')" temporaljson:"auth_type,omitzero,omitempty"`
	AuthConfig     TriggerAuthConfig     `json:"auth_config,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"auth_config,omitzero,omitempty"`
	Envelope       EventEnvelopeType     `json:"envelope" gorm:"notnull;<-:create;check:trigger_envelope_checker,envelope IN ('none','pubsub_push','cloudevents','sns')" temporaljson:"envelope,omitzero,omitempty"`
	TypeFrom       EventFieldSelector    `json:"type_from,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"type_from,omitzero,omitempty"`
	IDFrom         EventFieldSelector    `json:"id_from,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"id_from,omitzero,omitempty"`
	Status         TriggerStatus         `json:"status" gorm:"notnull;check:trigger_status_checker,status IN ('active','suspended')" temporaljson:"status,omitzero,omitempty"`
	LastEventAt    *time.Time            `json:"last_event_at,omitempty" temporaljson:"last_event_at,omitzero,omitempty"`
	Secrets        []TriggerSecret       `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	Events         []TriggerEvent        `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
}

func (e *Trigger) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "org_id_name_deleted_at"), Columns: []string{"org_id", "name", "deleted_at"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "ingress_key_hash"), Columns: []string{"ingress_key_hash"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
	}
}

func (e *Trigger) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewTriggerID()
	}
	if e.OrgID == "" {
		e.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if e.CreatedByID == "" {
		e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if e.AuthType == "" {
		e.AuthType = TriggerAuthTypeHMAC
	}
	if e.Envelope == "" {
		e.Envelope = EventEnvelopeTypeNone
	}
	if e.Status == "" {
		e.Status = TriggerStatusActive
	}
	return nil
}

type TriggerSecret struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID       string                `json:"-" gorm:"notnull;<-:create" temporaljson:"-"`
	Org         Org                   `json:"-" temporaljson:"-"`
	TriggerID   string                `json:"trigger_id" gorm:"notnull;<-:create" temporaljson:"trigger_id,omitzero,omitempty"`
	Trigger     Trigger               `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	KeyID       string                `json:"key_id" gorm:"notnull;<-:create" temporaljson:"key_id,omitzero,omitempty"`
	Secret      string                `json:"-" gorm:"notnull" temporaljson:"-"`
	NotBefore   time.Time             `json:"not_before" gorm:"notnull;<-:create" temporaljson:"not_before,omitzero,omitempty"`
	ExpiresAt   *time.Time            `json:"expires_at,omitempty" gorm:"check:trigger_secret_expiration_checker,expires_at IS NULL OR expires_at > not_before" temporaljson:"expires_at,omitzero,omitempty"`
	RevokedAt   *time.Time            `json:"revoked_at,omitempty" temporaljson:"revoked_at,omitzero,omitempty"`
	LastUsedAt  *time.Time            `json:"last_used_at,omitempty" temporaljson:"last_used_at,omitzero,omitempty"`
}

func (e *TriggerSecret) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "trigger_id_key_id"), Columns: []string{"trigger_id", "key_id"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "trigger_id"), Columns: []string{"trigger_id"}},
		{Name: indexes.Name(db, e, "inactive_secret_scrub"), Columns: []string{"id"}, Option: "WHERE secret <> '' AND (revoked_at IS NOT NULL OR expires_at IS NOT NULL)"},
	}
}

func (e *TriggerSecret) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewTriggerSecretID()
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

type TriggerEvent struct {
	ID                     string                  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt              time.Time               `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt              time.Time               `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt              soft_delete.DeletedAt   `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	TriggerID              string                  `json:"trigger_id" gorm:"notnull;<-:create" temporaljson:"trigger_id,omitzero,omitempty"`
	Trigger                Trigger                 `json:"-" gorm:"constraint:OnDelete:CASCADE" temporaljson:"-"`
	TriggerName            string                  `json:"trigger_name,omitempty" gorm:"-" temporaljson:"-"`
	TriggerSecretID        *string                 `json:"trigger_secret_id,omitempty" gorm:"<-:create" temporaljson:"trigger_secret_id,omitzero,omitempty"`
	TriggerSecret          TriggerSecret           `json:"-" temporaljson:"-"`
	OrgID                  string                  `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org                    Org                     `json:"-" temporaljson:"-"`
	ExternalID             string                  `json:"external_id" gorm:"notnull;<-:create" temporaljson:"external_id,omitzero,omitempty"`
	DedupeID               string                  `json:"-" gorm:"notnull;<-:create" temporaljson:"-"`
	Source                 string                  `json:"source,omitempty" gorm:"notnull;<-:create" temporaljson:"source,omitzero,omitempty"`
	EventType              string                  `json:"event_type" gorm:"notnull;<-:create" temporaljson:"event_type,omitzero,omitempty"`
	OccurredAt             *time.Time              `json:"occurred_at,omitempty" gorm:"<-:create" temporaljson:"occurred_at,omitzero,omitempty"`
	ReceivedAt             time.Time               `json:"received_at" gorm:"notnull;<-:create" temporaljson:"received_at,omitzero,omitempty"`
	Payload                json.RawMessage         `json:"payload" gorm:"serializer:json;type:jsonb;notnull;<-:create" temporaljson:"payload,omitzero,omitempty"`
	Headers                map[string][]string     `json:"headers,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"headers,omitzero,omitempty"`
	RawBody                []byte                  `json:"-" gorm:"type:bytea;notnull;<-:create" temporaljson:"-"`
	RawBodySHA256          string                  `json:"raw_body_sha256" gorm:"notnull;<-:create" temporaljson:"raw_body_sha256,omitzero,omitempty"`
	PayloadSHA256          string                  `json:"payload_sha256" gorm:"<-:create" temporaljson:"payload_sha256,omitzero,omitempty"`
	RawBodySize            int64                   `json:"raw_body_size" gorm:"<-:create" temporaljson:"raw_body_size,omitzero,omitempty"`
	RawContentType         string                  `json:"raw_content_type,omitempty" gorm:"<-:create" temporaljson:"raw_content_type,omitzero,omitempty"`
	PayloadContentType     string                  `json:"payload_content_type,omitempty" gorm:"<-:create" temporaljson:"payload_content_type,omitzero,omitempty"`
	SecretKeyID            string                  `json:"secret_key_id,omitempty" gorm:"<-:create" temporaljson:"secret_key_id,omitzero,omitempty"`
	RoutingStatus          EventRoutingStatus      `json:"routing_status" gorm:"notnull;check:event_routing_status_checker,routing_status IN ('accepted','routing','matched','ignored','rejected','routing_failed')" temporaljson:"routing_status,omitzero,omitempty"`
	RoutingGenerationToken *string                 `json:"-" temporaljson:"-"`
	RoutingError           string                  `json:"routing_error,omitempty" temporaljson:"routing_error,omitzero,omitempty"`
	RoutingStartedAt       *time.Time              `json:"routing_started_at,omitempty" temporaljson:"routing_started_at,omitzero,omitempty"`
	RoutingCompletedAt     *time.Time              `json:"routing_completed_at,omitempty" temporaljson:"routing_completed_at,omitzero,omitempty"`
	MatchCount             int                     `json:"match_count" temporaljson:"match_count,omitzero,omitempty"`
	WaiterMatchCount       int                     `json:"waiter_match_count" temporaljson:"waiter_match_count,omitzero,omitempty"`
	DispatchCount          int                     `json:"dispatch_count" temporaljson:"dispatch_count,omitzero,omitempty"`
	MatchExplanations      []TriggerRuleEvaluation `json:"match_explanations,omitempty" gorm:"serializer:json;type:jsonb" temporaljson:"match_explanations,omitzero,omitempty"`
	ExplanationsTruncated  bool                    `json:"explanations_truncated,omitempty" temporaljson:"explanations_truncated,omitzero,omitempty"`
}

func (e *TriggerEvent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "trigger_id_source_dedupe_id"), Columns: []string{"trigger_id", "source", "dedupe_id"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id_received_at_id"), Columns: []string{"org_id", "received_at", "id"}},
		{Name: indexes.Name(db, e, "org_id_trigger_id_received_at_id"), Columns: []string{"org_id", "trigger_id", "received_at", "id"}},
		{Name: indexes.Name(db, e, "routing_status_received_at_id"), Columns: []string{"routing_status", "received_at", "id"}},
	}
}
func (e *TriggerEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewTriggerEventID()
	}
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now()
	}
	if e.DedupeID == "" {
		e.DedupeID = e.ExternalID
	}
	if e.RoutingStatus == "" {
		e.RoutingStatus = EventRoutingStatusAccepted
	}
	return nil
}

type TriggerRule struct {
	ID            string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID   string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt     time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt     time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt     soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID         string                `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org           Org                   `json:"-" temporaljson:"-"`
	AppID         string                `json:"app_id" gorm:"notnull;<-:create" temporaljson:"app_id,omitzero,omitempty"`
	App           App                   `json:"-" temporaljson:"-"`
	AppConfigID   string                `json:"app_config_id" gorm:"notnull;<-:create" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig     AppConfig             `json:"-" temporaljson:"app_config,omitzero,omitempty"`
	TriggerID     string                `json:"trigger_id" gorm:"notnull;<-:create" temporaljson:"trigger_id,omitzero,omitempty"`
	Trigger       Trigger               `json:"-" temporaljson:"-"`
	Name          string                `json:"name" gorm:"notnull;<-:create" temporaljson:"name,omitzero,omitempty"`
	Enabled       bool                  `json:"enabled" gorm:"notnull;<-:create" temporaljson:"enabled,omitzero,omitempty"`
	SuspendedAt   *time.Time            `json:"suspended_at,omitempty" temporaljson:"suspended_at,omitzero,omitempty"`
	SuspendedByID *string               `json:"suspended_by_id,omitempty" temporaljson:"suspended_by_id,omitzero,omitempty"`
	ValidFrom     time.Time             `json:"valid_from" gorm:"notnull;<-:create" temporaljson:"valid_from,omitzero,omitempty"`
	ValidTo       *time.Time            `json:"valid_to,omitempty" gorm:"check:trigger_rule_validity_checker,valid_to IS NULL OR valid_to > valid_from" temporaljson:"valid_to,omitzero,omitempty"`
	EventTypes    pq.StringArray        `json:"event_types,omitempty" gorm:"type:text[];<-:create" temporaljson:"event_types,omitzero,omitempty"`
	Filters       []TriggerFilter       `json:"filters" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"filters,omitzero,omitempty"`
	TargetType    TriggerTargetType     `json:"target_type" gorm:"notnull;<-:create;check:trigger_rule_target_type_checker,target_type IN ('app_branch_run','runbook')" temporaljson:"target_type,omitzero,omitempty"`
	AppBranchID   *string               `json:"app_branch_id,omitempty" gorm:"<-:create;check:trigger_rule_target_shape_checker,(target_type = 'app_branch_run' AND app_branch_id IS NOT NULL AND runbook_id IS NULL AND install_name = '') OR (target_type = 'runbook' AND app_branch_id IS NULL AND runbook_id IS NOT NULL AND install_name <> '')" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch     AppBranch             `json:"-" temporaljson:"-"`
	RunbookID     *string               `json:"runbook_id,omitempty" gorm:"<-:create" temporaljson:"runbook_id,omitzero,omitempty"`
	InstallName   string                `json:"install_name,omitempty" gorm:"<-:create" temporaljson:"install_name,omitzero,omitempty"`
	InputMappings map[string]string     `json:"input_mappings,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"input_mappings,omitzero,omitempty"`
	Force         bool                  `json:"force" gorm:"<-:create" temporaljson:"force,omitzero,omitempty"`
	PlanOnly      bool                  `json:"plan_only" gorm:"<-:create" temporaljson:"plan_only,omitzero,omitempty"`
	ConfigHash    string                `json:"config_hash" gorm:"notnull;<-:create" temporaljson:"config_hash,omitzero,omitempty"`
}

func ActiveTriggerConfigIDs(configs []AppConfig) map[string]string {
	active := make(map[string]string)
	for i := range configs {
		config := &configs[i]
		if _, ok := active[config.AppID]; ok || config.Labels["source"] == string(AppBranchRunTypeGitPreview) {
			continue
		}
		active[config.AppID] = config.ID
	}
	return active
}

func (e *TriggerRule) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "app_config_id_name_deleted_at"), Columns: []string{"app_config_id", "name", "deleted_at"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, e, "trigger_id_valid_from"), Columns: []string{"trigger_id", "valid_from"}},
		{Name: indexes.Name(db, e, "trigger_id_app_config_id"), Columns: []string{"trigger_id", "app_config_id"}},
	}
}

func (e *TriggerRule) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = domains.NewTriggerRuleID()
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
	ID                 string              `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID        string              `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt          time.Time           `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	OrgID              string              `json:"org_id" gorm:"notnull;<-:create" temporaljson:"org_id,omitzero,omitempty"`
	Org                Org                 `json:"-" temporaljson:"-"`
	AppID              string              `json:"app_id" gorm:"notnull;<-:create" temporaljson:"app_id,omitzero,omitempty"`
	App                App                 `json:"-" temporaljson:"-"`
	TriggerEventID     string              `json:"trigger_event_id" gorm:"notnull;<-:create" temporaljson:"trigger_event_id,omitzero,omitempty"`
	TriggerEvent       TriggerEvent        `json:"-" temporaljson:"-"`
	TriggerRuleID      string              `json:"trigger_rule_id" gorm:"notnull;<-:create" temporaljson:"trigger_rule_id,omitzero,omitempty"`
	TriggerRule        TriggerRule         `json:"-" temporaljson:"-"`
	ReplayID           *string             `json:"replay_id,omitempty" gorm:"<-:create" temporaljson:"replay_id,omitzero,omitempty"`
	IdempotencyKey     string              `json:"idempotency_key" gorm:"notnull;<-:create" temporaljson:"idempotency_key,omitzero,omitempty"`
	TargetType         TriggerTargetType   `json:"target_type" gorm:"notnull;<-:create;check:event_dispatch_target_type_checker,target_type IN ('app_branch_run','runbook')" temporaljson:"target_type,omitzero,omitempty"`
	TargetID           string              `json:"target_id" gorm:"<-:create;check:event_dispatch_target_shape_checker,status IN ('dead_lettered','cancelled') OR target_id <> ''" temporaljson:"target_id,omitzero,omitempty"`
	RunbookConfigID    *string             `json:"runbook_config_id,omitempty" gorm:"<-:create" temporaljson:"runbook_config_id,omitzero,omitempty"`
	MappedInputs       map[string]string   `json:"mapped_inputs,omitempty" gorm:"serializer:json;type:jsonb;<-:create" temporaljson:"mapped_inputs,omitzero,omitempty"`
	Status             EventDispatchStatus `json:"status" gorm:"notnull;check:event_dispatch_status_checker,status IN ('pending','dispatching','triggered','retryable_failed','dead_lettered','cancelled')" temporaljson:"status,omitzero,omitempty"`
	Attempts           int                 `json:"attempts" gorm:"check:event_dispatch_attempts_checker,attempts >= 0" temporaljson:"attempts,omitzero,omitempty"`
	GenerationToken    string              `json:"-" temporaljson:"-"`
	ExecutionToken     string              `json:"-" temporaljson:"-"`
	NextAttemptAt      *time.Time          `json:"next_attempt_at,omitempty" temporaljson:"next_attempt_at,omitzero,omitempty"`
	Error              string              `json:"error,omitempty" temporaljson:"error,omitzero,omitempty"`
	QueueSignalID      *string             `json:"queue_signal_id,omitempty" temporaljson:"queue_signal_id,omitzero,omitempty"`
	ResultResourceType string              `json:"result_resource_type,omitempty" temporaljson:"result_resource_type,omitzero,omitempty"`
	ResultResourceID   string              `json:"result_resource_id,omitempty" temporaljson:"result_resource_id,omitzero,omitempty"`
	WorkflowID         string              `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
	StartedAt          *time.Time          `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`
	TriggeredAt        *time.Time          `json:"triggered_at,omitempty" temporaljson:"triggered_at,omitzero,omitempty"`
	FailedAt           *time.Time          `json:"failed_at,omitempty" temporaljson:"failed_at,omitzero,omitempty"`
}

func (e *EventDispatch) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, e, "idempotency_key"), Columns: []string{"idempotency_key"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}},
		{Name: indexes.Name(db, e, "trigger_event_id"), Columns: []string{"trigger_event_id"}},
		{Name: indexes.Name(db, e, "status_next_attempt_at"), Columns: []string{"status", "next_attempt_at"}},
		{Name: indexes.Name(db, e, "org_id_created_at_id"), Columns: []string{"org_id", "created_at", "id"}},
		{Name: indexes.Name(db, e, "org_id_trigger_event_id_created_at_id"), Columns: []string{"org_id", "trigger_event_id", "created_at", "id"}},
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
