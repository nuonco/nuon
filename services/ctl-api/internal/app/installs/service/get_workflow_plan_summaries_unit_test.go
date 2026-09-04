package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestPlanChangeCountsFixtures(t *testing.T) {
	tests := []struct {
		name         string
		approvalType app.WorkflowStepApprovalType
		expected     app.StepChangeCounts
	}{
		{
			name:         "terraform",
			approvalType: app.TerraformPlanApprovalType,
			expected:     app.StepChangeCounts{Create: 1, Update: 1, Delete: 1, Replace: 1},
		},
		{
			name:         "pulumi",
			approvalType: app.PulumiApprovalType,
			expected:     app.StepChangeCounts{Create: 2, Update: 1, Delete: 1, Replace: 1, Noop: 3},
		},
		{
			name:         "helm",
			approvalType: app.HelmApprovalApprovalType,
			expected:     app.StepChangeCounts{Create: 2, Update: 1, Delete: 1},
		},
		{
			name:         "kubernetes",
			approvalType: app.KubernetesManifestApprovalType,
			expected:     app.StepChangeCounts{Create: 1, Update: 1, Delete: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contents, err := os.Open(filepath.Join("testdata", "plan-summaries", test.name+".json"))
			require.NoError(t, err)
			defer contents.Close()

			counts, err := planChangeCounts(test.approvalType, contents)
			require.NoError(t, err)
			require.Equal(t, test.expected, counts)
		})
	}
}

func TestPlanChangeCounts(t *testing.T) {
	tests := []struct {
		name         string
		approvalType app.WorkflowStepApprovalType
		contents     string
		expected     app.StepChangeCounts
	}{
		{
			name:         "terraform",
			approvalType: app.TerraformPlanApprovalType,
			contents: `{
				"resource_changes": [
					{"change":{"actions":["create"]}},
					{"change":{"actions":["update"]}},
					{"change":{"actions":["delete"]}},
					{"change":{"actions":["delete","create"]}},
					{"change":{"actions":["create","delete"]}},
					{"change":{"actions":["no-op"]}},
					{"change":{"actions":["read"]}}
				]
			}`,
			expected: app.StepChangeCounts{
				Create: 1, Update: 1, Delete: 1, Replace: 2, Noop: 1,
			},
		},
		{
			name:         "pulumi change summary",
			approvalType: app.PulumiApprovalType,
			contents: `{
				"change_summary":{
					"create":3,
					"update":2,
					"delete":1,
					"replace":4,
					"same":5,
					"create-replacement":4,
					"delete-replaced":4
				}
			}`,
			expected: app.StepChangeCounts{
				Create: 3, Update: 2, Delete: 1, Replace: 4, Noop: 5,
			},
		},
		{
			name:         "pulumi resource fallback",
			approvalType: app.PulumiApprovalType,
			contents: `{
				"resource_changes":[
					{"action":"create"},
					{"action":"update"},
					{"action":"delete"},
					{"action":"replace"},
					{"action":"same"}
				]
			}`,
			expected: app.StepChangeCounts{
				Create: 1, Update: 1, Delete: 1, Replace: 1, Noop: 1,
			},
		},
		{
			name:         "helm plan summary",
			approvalType: app.HelmApprovalApprovalType,
			contents: `{
				"plan":"default, api, Deployment (apps) to be changed\nPlan: 2 to add, 3 to change, 4 to destroy.\n",
				"helm_content_diff":[]
			}`,
			expected: app.StepChangeCounts{Create: 2, Update: 3, Delete: 4},
		},
		{
			name:         "helm content fallback",
			approvalType: app.HelmApprovalApprovalType,
			contents: `{
				"plan":"",
				"helm_content_diff":[
					{"before":null,"after":{"kind":"ConfigMap"}},
					{"before":{"kind":"Deployment"},"after":{"kind":"Deployment"}},
					{"before":{"kind":"Service"},"after":null}
				]
			}`,
			expected: app.StepChangeCounts{Create: 1, Update: 1, Delete: 1},
		},
		{
			name:         "helm no diff",
			approvalType: app.HelmApprovalApprovalType,
			contents:     `{"plan":"","helm_content_diff":null}`,
			expected:     app.StepChangeCounts{},
		},
		{
			name:         "kubernetes",
			approvalType: app.KubernetesManifestApprovalType,
			contents: `{
				"k8s_content_diff":[
					{"op":"apply","type":2},
					{"op":"apply","type":3},
					{"op":"delete","type":2},
					{"op":"apply","type":1},
					{"op":"apply","type":0},
					{"op":"apply","type":3,"error":"dry run failed"}
				]
			}`,
			expected: app.StepChangeCounts{Create: 1, Update: 1, Delete: 2, Noop: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counts, err := planChangeCounts(test.approvalType, strings.NewReader(test.contents))
			require.NoError(t, err)
			require.Equal(t, test.expected, counts)
		})
	}
}

func TestPlanChangeCountsRejectsInvalidJSON(t *testing.T) {
	for _, approvalType := range []app.WorkflowStepApprovalType{
		app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType,
	} {
		t.Run(string(approvalType), func(t *testing.T) {
			_, err := planChangeCounts(approvalType, strings.NewReader(`{`))
			require.Error(t, err)
		})
	}
}

func TestNewStepChangeSummary(t *testing.T) {
	tests := []struct {
		name           string
		approvalType   app.WorkflowStepApprovalType
		stepStatus     app.Status
		expectedStatus app.StepChangeStatus
		hasDetail      bool
	}{
		{
			name:           "terraform awaiting approval",
			approvalType:   app.TerraformPlanApprovalType,
			stepStatus:     app.AwaitingApproval,
			expectedStatus: app.StepChangeStatusPendingApproval,
			hasDetail:      true,
		},
		{
			name:           "app branch approved",
			approvalType:   app.AppBranchPlanApprovalType,
			stepStatus:     app.WorkflowStepApprovalStatusApproved,
			expectedStatus: app.StepChangeStatusApproved,
			hasDetail:      true,
		},
		{
			name:           "install creation denied",
			approvalType:   app.InstallCreationApprovalType,
			stepStatus:     app.WorkflowStepApprovalStatusApprovalDenied,
			expectedStatus: app.StepChangeStatusDenied,
			hasDetail:      false,
		},
		{
			name:           "terraform applied",
			approvalType:   app.TerraformPlanApprovalType,
			stepStatus:     app.StatusSuccess,
			expectedStatus: app.StepChangeStatusApplied,
			hasDetail:      true,
		},
		{
			name:           "terraform generating",
			approvalType:   app.TerraformPlanApprovalType,
			stepStatus:     app.StatusInProgress,
			expectedStatus: app.StepChangeStatusGenerating,
			hasDetail:      true,
		},
		{
			name:           "terraform failed",
			approvalType:   app.TerraformPlanApprovalType,
			stepStatus:     app.StatusError,
			expectedStatus: app.StepChangeStatusError,
			hasDetail:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			step := app.WorkflowStep{
				ID:     "iws00000000000000000000000",
				Name:   "sync and plan api",
				Status: app.CompositeStatus{Status: test.stepStatus},
				Approval: &app.WorkflowStepApproval{
					ID:   "waa00000000000000000000000",
					Type: test.approvalType,
				},
			}

			summary := newStepChangeSummary(step, "api")
			require.Equal(t, step.ID, summary.StepID)
			require.Equal(t, step.Name, summary.StepName)
			require.Equal(t, step.Approval.ID, summary.ApprovalID)
			require.Equal(t, "api", summary.ComponentName)
			require.Equal(t, app.StepChangePlanType(test.approvalType), summary.PlanType)
			require.Equal(t, test.expectedStatus, summary.Status)
			require.Equal(t, test.hasDetail, summary.HasDetail)
			require.Equal(t, app.StepChangeCounts{}, summary.Counts)
		})
	}
}

func TestIsSummaryApprovalType(t *testing.T) {
	require.False(t, isSummaryApprovalType(app.NoopApprovalType))
	require.False(t, isSummaryApprovalType(app.ApproveAllApprovalType))
	require.True(t, isSummaryApprovalType(app.TerraformPlanApprovalType))
}

func TestPlanTypeHasCounts(t *testing.T) {
	require.True(t, planTypeHasCounts(app.TerraformPlanApprovalType))
	require.True(t, planTypeHasCounts(app.PulumiApprovalType))
	require.True(t, planTypeHasCounts(app.HelmApprovalApprovalType))
	require.True(t, planTypeHasCounts(app.KubernetesManifestApprovalType))
	require.False(t, planTypeHasCounts(app.AppBranchPlanApprovalType))
	require.False(t, planTypeHasCounts(app.InstallCreationApprovalType))
}
