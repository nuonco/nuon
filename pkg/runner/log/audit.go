package log

import (
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// AuditAttr tags records the bundled collector forwards to the customer's
// backend. Its filter drops anything where the log record attribute
// nuon.audit != "true", so the value must be this exact string and must be a
// record attribute rather than a resource attribute.
const (
	AuditAttr      = "nuon.audit"
	AuditAttrValue = "true"

	AuditEventAttr   = "nuon.audit.event"
	AuditOutcomeAttr = "nuon.audit.outcome"
)

const (
	OutcomeStarted   = "started"
	OutcomeSucceeded = "succeeded"
	OutcomeFailed    = "failed"
)

// auditEventTypes maps the runner job groups that produce customer-visible
// changes onto the event types used by the install_audit_logs view, so the
// streamed audit log stays consistent with the one customers download today.
var auditEventTypes = map[models.AppRunnerJobGroup]string{
	models.AppRunnerJobGroupDeploy:  "install_deploy",
	models.AppRunnerJobGroupActions: "install_action_workflow_run",
	models.AppRunnerJobGroupSandbox: "install_sandbox_run",
}

func AuditEventType(job *models.AppRunnerJob) string {
	if job == nil {
		return ""
	}
	return auditEventTypes[job.Group]
}

func IsAuditable(job *models.AppRunnerJob) bool {
	return AuditEventType(job) != ""
}

// NewAudit derives an audit logger from a job logger. It returns nil for jobs
// whose group is not auditable, so callers must nil-check before emitting;
// tagging every job would send the runner's entire log volume to the customer.
func NewAudit(l *zap.Logger, job *models.AppRunnerJob) *zap.Logger {
	if l == nil || !IsAuditable(job) {
		return nil
	}

	fields := []zap.Field{
		zap.String(AuditAttr, AuditAttrValue),
		zap.String(AuditEventAttr, AuditEventType(job)),
	}
	fields = append(fields, jobs.AuditFields(job)...)

	return l.With(fields...)
}

// AuditEvent emits one audit record, and does nothing when l is nil so callers
// can pass the result of NewAudit straight through without branching on
// whether the job was auditable.
func AuditEvent(l *zap.Logger, msg, outcome string, fields ...zap.Field) {
	if l == nil {
		return
	}
	l.Info(msg, append([]zap.Field{zap.String(AuditOutcomeAttr, outcome)}, fields...)...)
}
