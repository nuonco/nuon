package jobs

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// auditMetadataAttrs maps runner job metadata keys, which ctl-api populates at
// job creation and vary by job group, onto the attribute names we emit.
var auditMetadataAttrs = map[string]string{
	"install_id":             "install.id",
	"flow_install_id":        "flow_install.id",
	"flow_workflow_id":       "flow_workflow.id",
	"app_id":                 "app.id",
	"component_id":           "component.id",
	"component_name":         "component.name",
	"action_workflow_name":   "action_workflow.name",
	"action_workflow_id":     "action_workflow.id",
	"action_workflow_run_id": "action_workflow_run.id",
	"deploy_id":              "deploy.id",
	"sandbox_run_id":         "sandbox_run.id",
	"sandbox_run_type":       "sandbox_run.type",
}

// auditPairs is the single source of truth for the audit envelope. Both the
// zap field and OTEL attribute forms are derived from it so log records and
// spans never drift apart.
func auditPairs(job *models.AppRunnerJob) [][2]string {
	if job == nil {
		return nil
	}

	pairs := [][2]string{
		{string(semconv.UserIDKey), job.CreatedByID},
		{"runner_job.created_by_id", job.CreatedByID},
		{"runner_job.group", string(job.Group)},
		{"runner_job.operation", string(job.Operation)},
		{"runner_job.executor", string(job.Executor)},
		{"runner_job.owner_id", job.OwnerID},
		{"runner_job.owner_type", job.OwnerType},
		{"org.id", job.OrgID},
		{"runner_process.id", job.RunnerProcessID},
	}

	for metaKey, attrKey := range auditMetadataAttrs {
		if v, ok := job.Metadata[metaKey]; ok {
			pairs = append(pairs, [2]string{attrKey, v})
		}
	}

	out := make([][2]string, 0, len(pairs))
	for _, p := range pairs {
		if p[1] != "" {
			out = append(out, p)
		}
	}
	return out
}

func AuditFields(job *models.AppRunnerJob) []zap.Field {
	pairs := auditPairs(job)
	fields := make([]zap.Field, 0, len(pairs))
	for _, p := range pairs {
		fields = append(fields, zap.String(p[0], p[1]))
	}
	return fields
}

// AuditMetadata builds the ctx-threaded job metadata that op.Start stamps onto
// every descendant span. It carries only the dimensions worth querying spans
// by; the full envelope rides on the job logger instead.
func AuditMetadata(job *models.AppRunnerJob, executionID, stepName string) pkgctx.JobMetadata {
	meta := pkgctx.JobMetadata{
		RunnerJobExecutionID: executionID,
		StepName:             stepName,
	}
	if job == nil {
		return meta
	}

	meta.RunnerJobID = job.ID
	meta.JobGroup = string(job.Group)
	meta.JobOperation = string(job.Operation)
	meta.Executor = string(job.Executor)
	meta.OrgID = job.OrgID
	meta.InstallID = job.Metadata["install_id"]
	return meta
}

func AuditAttrs(job *models.AppRunnerJob) []attribute.KeyValue {
	pairs := auditPairs(job)
	attrs := make([]attribute.KeyValue, 0, len(pairs))
	for _, p := range pairs {
		attrs = append(attrs, attribute.String(p[0], p[1]))
	}
	return attrs
}
