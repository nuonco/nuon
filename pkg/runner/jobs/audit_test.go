package jobs

import (
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func fieldMap(job *models.AppRunnerJob) map[string]string {
	out := map[string]string{}
	for _, p := range auditPairs(job) {
		out[p[0]] = p[1]
	}
	return out
}

func TestAuditPairsPopulatedJob(t *testing.T) {
	job := &models.AppRunnerJob{
		CreatedByID:     "acct123",
		Group:           models.AppRunnerJobGroupDeploy,
		Operation:       models.AppRunnerJobOperationTypeApplyDashPlan,
		Executor:        "org-runner",
		OwnerID:         "dep123",
		OwnerType:       "install_deploys",
		OrgID:           "org123",
		RunnerProcessID: "proc123",
		Metadata: map[string]string{
			"install_id":           "inst123",
			"component_name":       "api",
			"deploy_id":            "dep123",
			"flow_workflow_id":     "wf123",
			"install_component_id": "instcmp123",
			"notebook_cell_run_id": "cellrun123",
		},
	}

	want := map[string]string{
		"user.id":                  "acct123",
		"runner_job.created_by_id": "acct123",
		"runner_job.group":         "deploy",
		"runner_job.operation":     "apply-plan",
		"runner_job.executor":      "org-runner",
		"runner_job.owner_id":      "dep123",
		"runner_job.owner_type":    "install_deploys",
		"org.id":                   "org123",
		"runner_process.id":        "proc123",
		"install.id":               "inst123",
		"component.name":           "api",
		"deploy.id":                "dep123",
		"flow_workflow.id":         "wf123",
		"install_component.id":     "instcmp123",
		"notebook_cell_run.id":     "cellrun123",
	}

	got := fieldMap(job)
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d attributes, want %d: %v", len(got), len(want), got)
	}

	if n := len(AuditFields(job)); n != len(want) {
		t.Errorf("AuditFields returned %d fields, want %d", n, len(want))
	}
	if n := len(AuditAttrs(job)); n != len(want) {
		t.Errorf("AuditAttrs returned %d attrs, want %d", n, len(want))
	}
}

func TestAuditPairsSkipsEmptyValues(t *testing.T) {
	job := &models.AppRunnerJob{
		Group:    models.AppRunnerJobGroupDeploy,
		Metadata: map[string]string{"install_id": "", "app_id": "app123"},
	}

	got := fieldMap(job)
	for _, key := range []string{"user.id", "org.id", "install.id", "runner_job.owner_id"} {
		if _, ok := got[key]; ok {
			t.Errorf("%s present for an empty value", key)
		}
	}
	if got["app.id"] != "app123" {
		t.Errorf("app.id = %q, want %q", got["app.id"], "app123")
	}
}

func TestAuditPairsEmptyJob(t *testing.T) {
	if n := len(AuditFields(&models.AppRunnerJob{})); n != 0 {
		t.Errorf("AuditFields on an empty job returned %d fields, want 0", n)
	}
	if n := len(AuditFields(nil)); n != 0 {
		t.Errorf("AuditFields(nil) returned %d fields, want 0", n)
	}
	if n := len(AuditAttrs(nil)); n != 0 {
		t.Errorf("AuditAttrs(nil) returned %d attrs, want 0", n)
	}
}
