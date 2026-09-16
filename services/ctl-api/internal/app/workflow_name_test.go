package app

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// every type declared in workflow.go — add new ones here so the titles below
// stay mandatory rather than silently falling back to the lowercased type
var allWorkflowTypes = []WorkflowType{
	WorkflowTypeProvision,
	WorkflowTypeDeprovision,
	WorkflowTypeDeprovisionSandbox,
	WorkflowTypeManualDeploy,
	WorkflowTypeInputUpdate,
	WorkflowTypeDeployComponents,
	WorkflowTypeTeardownComponent,
	WorkflowTypeTeardownComponents,
	WorkflowTypeReprovisionSandbox,
	WorkflowTypeDriftRunReprovisionSandbox,
	WorkflowTypeActionWorkflowRun,
	WorkflowTypeSyncSecrets,
	WorkflowTypeDriftRun,
	WorkflowTypeAppBranchesRun,
	WorkflowTypeAppBranchesConfigRepoUpdate,
	WorkflowTypeAppBranchesComponentRepoUpdate,
	WorkflowTypeAppBranchConfigUpdate,
	WorkflowTypeReprovision,
	WorkflowTypeReprovisionStack,
	WorkflowTypeAppConfigBuild,
	WorkflowTypeRunbookRun,
	WorkflowTypeComponentEnabled,
	WorkflowTypeComponentDisabled,
}

func TestWorkflowTypeTitlesAreExhaustive(t *testing.T) {
	for _, wt := range allWorkflowTypes {
		// app_branches_manual_update is titled from its metadata instead
		if wt == WorkflowTypeAppBranchesRun {
			continue
		}
		if wt.Name() == "" {
			t.Errorf("%s has no Name()", wt)
		}
		if wt.PastTenseName() == "" {
			t.Errorf("%s has no PastTenseName()", wt)
		}
	}
}

func TestComputeNameNeverFallsBackToRawType(t *testing.T) {
	for _, wt := range allWorkflowTypes {
		name := (&Workflow{Type: wt}).ComputeName()
		if name == "" {
			t.Errorf("%s computed an empty name", wt)
		}
		if name == string(wt) {
			t.Errorf("%s computed the raw type as its name", wt)
		}
	}
}

func TestAppBranchRunName(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		expected string
	}{
		{
			name:     "manual run",
			metadata: map[string]string{"event_type": "manual"},
			expected: "Run",
		},
		{
			name: "manual run with commit",
			metadata: map[string]string{
				"event_type": "manual",
				"commit_sha": "abcdef1234567890",
			},
			expected: "Run",
		},
		{
			name: "pull request",
			metadata: map[string]string{
				"event_type": "pull_request",
				"pr_number":  "42",
				"commit_sha": "abcdef1234567890",
			},
			expected: "PR #42",
		},
		{
			name: "push",
			metadata: map[string]string{
				"event_type": "push",
				"commit_sha": "abcdef1234567890",
			},
			expected: "VCS push",
		},
		{
			name:     "onboarding",
			metadata: map[string]string{"event_type": "onboarding"},
			expected: "Onboarding run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := make(pgtype.Hstore, len(tt.metadata))
			for key, value := range tt.metadata {
				value := value
				metadata[key] = &value
			}
			workflow := &Workflow{
				Type:     WorkflowTypeAppBranchesRun,
				Metadata: metadata,
			}

			require.Equal(t, tt.expected, workflow.ComputeName())
		})
	}
}
