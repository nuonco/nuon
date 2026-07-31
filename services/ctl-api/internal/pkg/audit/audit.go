package audit

import (
	"context"

	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// Attr tags records the gateway collector forwards to the customer's backend.
// Collector filters compare the log record attribute against the string
// "true", so the value must not be emitted as a bool.
const (
	Attr        = "nuon.audit"
	AttrValue   = "true"
	EventAttr   = "nuon.audit.event"
	OutcomeAttr = "nuon.audit.outcome"
)

// EventType mirrors the type column of the install_audit_logs view so streamed
// audit records stay consistent with the ones customers download today.
type EventType string

const (
	EventInstallDeploy            EventType = "install_deploy"
	EventInstallActionWorkflowRun EventType = "install_action_workflow_run"
	EventInstallSandboxRun        EventType = "install_sandbox_run"
	EventPolicyReport             EventType = "policy_report"
)

type Outcome string

const (
	OutcomeStarted   Outcome = "started"
	OutcomeSucceeded Outcome = "succeeded"
	OutcomeFailed    Outcome = "failed"
)

type Event struct {
	Type        EventType
	Message     string
	Outcome     Outcome
	InstallID   string
	AppID       string
	ComponentID string
	SubjectID   string
	SubjectType string
	Attrs       map[string]string
}

type Emitter struct {
	l *zap.Logger
}

func NewEmitter(l *zap.Logger) *Emitter {
	return &Emitter{l: l}
}

// Emit writes one audit record. The actor and org come from the same context
// that BeforeCreate uses to populate created_by_id, so an audit record can
// never disagree with the row it describes about who caused it.
func (e *Emitter) Emit(ctx context.Context, ev Event) {
	if e == nil || e.l == nil {
		return
	}

	fields := []zap.Field{
		zap.String(Attr, AttrValue),
		zap.String(EventAttr, string(ev.Type)),
		zap.String(OutcomeAttr, string(ev.Outcome)),
	}

	for key, value := range map[string]string{
		string(semconv.UserIDKey): keys.CreatedByIDFromContext(ctx),
		"org.id":                  keys.OrgIDFromContext(ctx),
		"install.id":              ev.InstallID,
		"app.id":                  ev.AppID,
		"component.id":            ev.ComponentID,
		"nuon.audit.subject_id":   ev.SubjectID,
		"nuon.audit.subject_type": ev.SubjectType,
	} {
		if value != "" {
			fields = append(fields, zap.String(key, value))
		}
	}
	for key, value := range ev.Attrs {
		if value != "" {
			fields = append(fields, zap.String(key, value))
		}
	}

	e.l.Info(ev.Message, fields...)
}
