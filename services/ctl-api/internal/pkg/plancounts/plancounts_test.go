package plancounts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCountsFixtures(t *testing.T) {
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
			contents, err := os.ReadFile(filepath.Join("testdata", test.name+".json"))
			require.NoError(t, err)

			counts, state, err := Counts(test.approvalType, string(contents))
			require.NoError(t, err)
			require.Equal(t, app.StepChangeStateOK, state)
			require.Equal(t, test.expected, counts)
		})
	}
}

func TestCounts(t *testing.T) {
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
			expected: app.StepChangeCounts{Create: 1, Update: 1, Delete: 1, Replace: 2, Noop: 1},
		},
		{
			name:         "terraform no changes",
			approvalType: app.TerraformPlanApprovalType,
			contents:     `{"resource_changes":[]}`,
			expected:     app.StepChangeCounts{},
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
			expected: app.StepChangeCounts{Create: 3, Update: 2, Delete: 1, Replace: 4, Noop: 5},
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
			expected: app.StepChangeCounts{Create: 1, Update: 1, Delete: 1, Replace: 1, Noop: 1},
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
		{
			name:         "kubernetes noop deploy",
			approvalType: app.KubernetesManifestApprovalType,
			contents:     `{"op":"noop"}`,
			expected:     app.StepChangeCounts{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			counts, state, err := Counts(test.approvalType, test.contents)
			require.NoError(t, err)
			require.Equal(t, app.StepChangeStateOK, state)
			require.Equal(t, test.expected, counts)
		})
	}
}

func TestCountsUnsupportedTypes(t *testing.T) {
	for _, approvalType := range []app.WorkflowStepApprovalType{
		app.NoopApprovalType,
		app.ApproveAllApprovalType,
		app.AppBranchPlanApprovalType,
		app.InstallCreationApprovalType,
	} {
		t.Run(string(approvalType), func(t *testing.T) {
			counts, state, err := Counts(approvalType, `{"anything":true}`)
			require.NoError(t, err)
			require.Equal(t, app.StepChangeStateUnsupported, state)
			require.Equal(t, app.StepChangeCounts{}, counts)
		})
	}
}

func TestCountsInvalidPlan(t *testing.T) {
	for _, approvalType := range []app.WorkflowStepApprovalType{
		app.TerraformPlanApprovalType,
		app.PulumiApprovalType,
		app.HelmApprovalApprovalType,
		app.KubernetesManifestApprovalType,
	} {
		t.Run(string(approvalType), func(t *testing.T) {
			counts, state, err := Counts(approvalType, `{`)
			require.Error(t, err)
			require.Equal(t, app.StepChangeStateError, state)
			require.Equal(t, app.StepChangeCounts{}, counts)
		})

		t.Run(string(approvalType)+" empty", func(t *testing.T) {
			_, state, err := Counts(approvalType, "  ")
			require.Error(t, err)
			require.Equal(t, app.StepChangeStateError, state)
		})
	}
}

func TestHasChanges(t *testing.T) {
	require.False(t, app.StepChangeCounts{}.HasChanges())
	require.False(t, app.StepChangeCounts{Noop: 12}.HasChanges())
	require.True(t, app.StepChangeCounts{Update: 1}.HasChanges())
	require.True(t, app.StepChangeCounts{Replace: 1}.HasChanges())
}
