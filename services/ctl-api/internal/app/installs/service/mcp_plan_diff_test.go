package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func TestSummarizeMCPWorkflowReturnsEveryPendingApproval(t *testing.T) {
	workflow := app.Workflow{ID: "wf"}
	workflow.Steps = []app.WorkflowStep{
		{
			ID:   "step-1",
			Name: "sandbox plan",
			Approval: &app.WorkflowStepApproval{
				ID:            "approval-1",
				Type:          app.TerraformPlanApprovalType,
				ChangesState:  app.StepChangeStateOK,
				ChangesCreate: 1,
			},
		},
		{
			ID:   "step-2",
			Name: "component plan",
			Approval: &app.WorkflowStepApproval{
				ID:            "approval-2",
				Type:          app.HelmApprovalApprovalType,
				ChangesState:  app.StepChangeStateOK,
				ChangesUpdate: 2,
			},
		},
	}

	summary := summarizeMCPWorkflow(workflow)

	require.Len(t, summary.PendingApprovals, 2)
	require.Equal(t, "approval-1", summary.PendingApprovals[0].ApprovalID)
	require.Equal(t, "step-1", summary.PendingApprovals[0].StepID)
	require.Equal(t, 1, summary.PendingApprovals[0].ChangesCreate)
	require.Equal(t, "approval-2", summary.PendingApprovals[1].ApprovalID)
	require.Equal(t, 2, summary.PendingApprovals[1].ChangesUpdate)
}

func TestBuildApprovalPlanDiffRedactsTerraformSensitiveValues(t *testing.T) {
	contents := `{
		"resource_changes": [{
			"address": "aws_iam_role.this",
			"type": "aws_iam_role",
			"name": "this",
			"change": {
				"actions": ["delete", "create"],
				"before": {"name": "old", "secret": "hunter2"},
				"after": {"name": "new", "secret": "hunter3"},
				"before_sensitive": {"secret": true},
				"after_sensitive": {"secret": true}
			}
		}]
	}`

	built, err := buildApprovalPlanDiff(app.TerraformPlanApprovalType, contents, "", "")
	require.NoError(t, err)
	require.Len(t, built.Changes, 1)
	require.Equal(t, "replace", built.Changes[0].Action)
	require.NotContains(t, built.Changes[0].Diff, "hunter2")
	require.NotContains(t, built.Changes[0].Diff, "hunter3")
	require.Contains(t, built.Changes[0].Diff, "~ secret: (sensitive value)")
	require.Contains(t, built.Changes[0].Diff, `+ name: "new"`)
}

func TestBuildApprovalPlanDiffPagesHelmChanges(t *testing.T) {
	contents := `{
		"helm_content_diff": [
			{"kind":"Deployment","name":"a","namespace":"ns","type":2,"entries":[{"type":2,"payload":"replicas: 2"}]},
			{"kind":"Deployment","name":"b","namespace":"ns","type":3,"entries":[{"type":1,"path":"spec.replicas","original":"1"},{"type":2,"path":"spec.replicas","applied":"2"}]},
			{"kind":"Deployment","name":"c","namespace":"ns","type":1,"entries":[{"type":1,"payload":"old"}]}
		]
	}`

	built, err := buildApprovalPlanDiff(app.HelmApprovalApprovalType, contents, "", "")
	require.NoError(t, err)
	require.Len(t, built.Changes, 3)
	require.Equal(t, []string{"create", "update", "delete"}, []string{
		built.Changes[0].Action,
		built.Changes[1].Action,
		built.Changes[2].Action,
	})

	page, hasMore := pagePlanDiffs(built.Changes, 2, 0)
	require.True(t, hasMore)
	require.Len(t, page, 2)
	require.Equal(t, 2, apiPkg.MCPNextOffset(0, 2, hasMore))
	require.Contains(t, built.Changes[1].Diff, "- spec.replicas:")
	require.Contains(t, built.Changes[1].Diff, "+ spec.replicas:")
}

func TestBuildApprovalPlanDiffKubernetesActions(t *testing.T) {
	contents := `{
		"k8s_content_diff": [
			{"kind":"ConfigMap","name":"app","namespace":"ns","op":"apply","type":2,"entries":[{"type":2,"payload":"key: value"}]},
			{"kind":"Secret","name":"app","namespace":"ns","op":"delete","type":1,"entries":[{"type":1,"payload":"secret"}]}
		]
	}`

	built, err := buildApprovalPlanDiff(app.KubernetesManifestApprovalType, contents, "", "")
	require.NoError(t, err)
	require.Equal(t, "create", built.Changes[0].Action)
	require.Equal(t, "ns/ConfigMap/app", built.Changes[0].Address)
	require.Equal(t, "delete", built.Changes[1].Action)
}

func TestBuildApprovalPlanDiffTooLargeSkipsParsing(t *testing.T) {
	contents := `{"resource_changes":[{"address":"aws_s3_bucket.this","type":"aws_s3_bucket","change":{"actions":["create"],"after":{"name":"` + strings.Repeat("a", mcpPlanMaxBytes) + `"}}}]}`

	built, err := buildApprovalPlanDiff(app.TerraformPlanApprovalType, contents, "", "")
	require.NoError(t, err)
	require.True(t, built.TooLarge)
	require.Empty(t, built.Changes)
	require.NotContains(t, strings.Join([]string{built.ChangesState}, ""), "aws_s3_bucket")
}

func TestBuildApprovalPlanDiffInstallCreationUnsupported(t *testing.T) {
	built, err := buildApprovalPlanDiff(app.InstallCreationApprovalType, `{"installs":["one"]}`, "", "")
	require.NoError(t, err)
	require.Equal(t, string(app.StepChangeStateUnsupported), built.ChangesState)
	require.Empty(t, built.Changes)
}

func TestBuildApprovalPlanDiffAppBranchSummary(t *testing.T) {
	contents := `{
		"install_group": "prod",
		"installs": [{
			"install_id": "inst",
			"install_name": "acme-prod",
			"diff": {
				"added": [{"component_name": "api"}],
				"removed": [{"component_name": "worker"}],
				"changed": [],
				"unchanged": [],
				"sandbox_changed": true,
				"stack_changed": false
			}
		}]
	}`

	built, err := buildApprovalPlanDiff(app.AppBranchPlanApprovalType, contents, "", "")
	require.NoError(t, err)
	require.Len(t, built.Changes, 1)
	require.Equal(t, "update", built.Changes[0].Action)
	require.Equal(t, "acme-prod", built.Changes[0].Address)
	require.Contains(t, built.Changes[0].Diff, "+ added api")
	require.Contains(t, built.Changes[0].Diff, "- removed worker")
	require.Contains(t, built.Changes[0].Diff, "sandbox changed")
}
