package log

import (
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuditEventType(t *testing.T) {
	tests := []struct {
		group models.AppRunnerJobGroup
		want  string
	}{
		{models.AppRunnerJobGroupDeploy, "install_deploy"},
		{models.AppRunnerJobGroupActions, "install_action_workflow_run"},
		{models.AppRunnerJobGroupSandbox, "install_sandbox_run"},
		{models.AppRunnerJobGroupHealthDashChecks, ""},
		{models.AppRunnerJobGroupSync, ""},
		{models.AppRunnerJobGroupBuild, ""},
		{models.AppRunnerJobGroupRunner, ""},
		{models.AppRunnerJobGroupOperations, ""},
		{models.AppRunnerJobGroupManagement, ""},
		{models.AppRunnerJobGroupEmpty, ""},
	}

	for _, tt := range tests {
		job := &models.AppRunnerJob{Group: tt.group}

		if got := AuditEventType(job); got != tt.want {
			t.Errorf("AuditEventType(%q) = %q, want %q", tt.group, got, tt.want)
		}
		if got, want := IsAuditable(job), tt.want != ""; got != want {
			t.Errorf("IsAuditable(%q) = %v, want %v", tt.group, got, want)
		}
	}
}

func TestAuditEventTypeNilJob(t *testing.T) {
	if got := AuditEventType(nil); got != "" {
		t.Errorf("AuditEventType(nil) = %q, want empty", got)
	}
	if IsAuditable(nil) {
		t.Error("IsAuditable(nil) = true, want false")
	}
}

func TestNewAuditReturnsNilForNonAuditableJob(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	l := zap.New(core)
	for _, group := range []models.AppRunnerJobGroup{
		models.AppRunnerJobGroupBuild,
		models.AppRunnerJobGroupOperations,
		models.AppRunnerJobGroupEmpty,
	} {
		if got := NewAudit(l, &models.AppRunnerJob{Group: group}); got != nil {
			t.Errorf("NewAudit(%q) returned a logger, want nil", group)
		}
	}
}

// The bundled collector drops records whose nuon.audit log-record attribute is
// not exactly the string "true". A bool value, or moving the attribute to the
// resource, silently discards every audit record, so pin both here.
func TestAuditAttrMatchesCollectorFilter(t *testing.T) {
	if AuditAttrValue != "true" {
		t.Fatalf("AuditAttrValue = %q, want \"true\"", AuditAttrValue)
	}

	core, logs := observer.New(zapcore.InfoLevel)
	l := NewAudit(zap.New(core), &models.AppRunnerJob{
		Group:       models.AppRunnerJobGroupDeploy,
		CreatedByID: "acct123",
		OrgID:       "org123",
		Metadata:    map[string]string{"install_id": "inst123"},
	})
	if l == nil {
		t.Fatal("NewAudit returned nil for a deploy job")
	}
	l.Info("deploy started")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("got %d log entries, want 1", len(entries))
	}

	fields := entries[0].ContextMap()
	got, ok := fields[AuditAttr]
	if !ok {
		t.Fatalf("%s not present on the record; fields: %v", AuditAttr, fields)
	}
	if s, isString := got.(string); !isString || s != "true" {
		t.Errorf("%s = %#v, want the string \"true\" (a bool is dropped by the collector filter)", AuditAttr, got)
	}

	for key, want := range map[string]string{
		AuditEventAttr:             "install_deploy",
		"user.id":                  "acct123",
		"runner_job.created_by_id": "acct123",
		"org.id":                   "org123",
		"install.id":               "inst123",
	} {
		if fields[key] != want {
			t.Errorf("%s = %v, want %q", key, fields[key], want)
		}
	}
}
