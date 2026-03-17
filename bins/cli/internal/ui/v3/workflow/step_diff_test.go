package workflow

import (
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestStepHasPlanDiff(t *testing.T) {
	tests := []struct {
		name         string
		workflowType models.AppWorkflowType
		step         *models.AppWorkflowStep
		want         bool
	}{
		{
			name: "supports plural install deploy target",
			step: &models.AppWorkflowStep{StepTargetType: "install_deploys", Name: "apply component"},
			want: true,
		},
		{
			name: "supports legacy singular install deploy target",
			step: &models.AppWorkflowStep{StepTargetType: "install_deploy", Name: "apply component"},
			want: true,
		},
		{
			name: "supports sync and plan step by name",
			step: &models.AppWorkflowStep{StepTargetType: "other", Name: "sync and plan api"},
			want: true,
		},
		{
			name: "does not show diff for system image sync steps",
			step: &models.AppWorkflowStep{
				ExecutionType:  "system",
				StepTargetType: "install_deploys",
				Name:           "sync img_nuon_dashboard_ui",
			},
			want: false,
		},
		{
			name: "still shows diff for non-system image sync names",
			step: &models.AppWorkflowStep{
				ExecutionType:  "hidden",
				StepTargetType: "install_deploys",
				Name:           "sync img_nuon_dashboard_ui",
			},
			want: true,
		},
		{
			name: "ignores unrelated steps",
			step: &models.AppWorkflowStep{StepTargetType: "other", Name: "await install stack"},
			want: false,
		},
		{
			name:         "shows diff for drift_check when status metadata plan_only is true",
			workflowType: workflowTypeDriftCheck,
			step: &models.AppWorkflowStep{
				StepTargetType: "other",
				Name:           "unrelated",
				Status: &models.AppCompositeStatus{Metadata: map[string]any{
					"plan_only": true,
				}},
			},
			want: true,
		},
		{
			name:         "supports string true for drift_check plan_only metadata",
			workflowType: workflowTypeDriftCheck,
			step: &models.AppWorkflowStep{
				StepTargetType: "other",
				Name:           "unrelated",
				Status: &models.AppCompositeStatus{Metadata: map[string]any{
					"plan_only": "true",
				}},
			},
			want: true,
		},
		{
			name:         "does not show diff for drift_check when plan_only metadata is false",
			workflowType: workflowTypeDriftCheck,
			step: &models.AppWorkflowStep{
				StepTargetType: "install_deploys",
				Name:           "sync and plan api",
				Status: &models.AppCompositeStatus{Metadata: map[string]any{
					"plan_only": false,
				}},
			},
			want: false,
		},
		{
			name:         "does not show diff for drift_check when plan_only metadata is missing",
			workflowType: workflowTypeDriftCheck,
			step: &models.AppWorkflowStep{
				StepTargetType: "install_deploys",
				Name:           "sync and plan api",
				Status:         &models.AppCompositeStatus{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{}
			if tt.workflowType != "" {
				m.workflow = &models.AppWorkflow{Type: tt.workflowType}
			}

			got := m.stepHasPlanDiff(tt.step)
			if got != tt.want {
				t.Fatalf("stepHasPlanDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}
