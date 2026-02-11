package plandiff

import (
	"testing"
)

func TestDetectPlanType(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		wantType    PlanType
		wantErr     bool
		errContains string
	}{
		{
			name:     "empty string",
			jsonStr:  "",
			wantType: PlanTypeUnknown,
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			jsonStr:  "{invalid json}",
			wantType: PlanTypeUnknown,
			wantErr:  true,
		},
		{
			name:     "empty JSON object",
			jsonStr:  "{}",
			wantType: PlanTypeUnknown,
			wantErr:  true,
		},
		{
			name:     "terraform plan with resource_changes",
			jsonStr:  `{"resource_changes": []}`,
			wantType: PlanTypeTerraform,
			wantErr:  false,
		},
		{
			name: "terraform plan with resource and output changes",
			jsonStr: `{
				"resource_changes": [{"address": "aws_instance.example"}],
				"output_changes": {"example": {"actions": ["create"]}}
			}`,
			wantType: PlanTypeTerraform,
			wantErr:  false,
		},
		{
			name:     "helm plan with helm_content_diff",
			jsonStr:  `{"helm_content_diff": []}`,
			wantType: PlanTypeHelm,
			wantErr:  false,
		},
		{
			name: "helm plan with content",
			jsonStr: `{
				"plan": "upgrade",
				"op": "install",
				"helm_content_diff": [{"kind": "Deployment", "name": "test"}]
			}`,
			wantType: PlanTypeHelm,
			wantErr:  false,
		},
		{
			name:     "kubernetes plan with k8s_content_diff",
			jsonStr:  `{"k8s_content_diff": []}`,
			wantType: PlanTypeKubernetes,
			wantErr:  false,
		},
		{
			name: "kubernetes plan with content",
			jsonStr: `{
				"plan": "apply",
				"op": "create",
				"k8s_content_diff": [{"kind": "Deployment", "name": "test"}]
			}`,
			wantType: PlanTypeKubernetes,
			wantErr:  false,
		},
		{
			name:     "unknown plan type",
			jsonStr:  `{"some_other_field": "value"}`,
			wantType: PlanTypeUnknown,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, _, err := DetectPlanType(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectPlanType() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DetectPlanType() unexpected error: %v", err)
				return
			}

			if gotType != tt.wantType {
				t.Errorf("DetectPlanType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestTryTerraform(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "valid terraform plan",
			jsonStr: `{"resource_changes": [{"address": "aws_instance.example", "type": "aws_instance", "name": "example", "change": {"actions": ["create"]}}]}`,
			wantErr: false,
		},
		{
			name:    "terraform plan with null resource_changes",
			jsonStr: `{"resource_changes": null}`,
			wantErr: true,
		},
		{
			name:    "not a terraform plan",
			jsonStr: `{"helm_content_diff": []}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			jsonStr: `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planType, plan, err := tryTerraform(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("tryTerraform() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("tryTerraform() unexpected error: %v", err)
				return
			}

			if planType != PlanTypeTerraform {
				t.Errorf("tryTerraform() type = %v, want %v", planType, PlanTypeTerraform)
			}

			if plan == nil {
				t.Error("tryTerraform() plan is nil")
			}
		})
	}
}

func TestTryHelm(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "valid helm plan",
			jsonStr: `{"helm_content_diff": [{"kind": "Deployment", "name": "test"}]}`,
			wantErr: false,
		},
		{
			name:    "helm plan with null helm_content_diff",
			jsonStr: `{"helm_content_diff": null}`,
			wantErr: true,
		},
		{
			name:    "not a helm plan",
			jsonStr: `{"resource_changes": []}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planType, plan, err := tryHelm(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("tryHelm() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("tryHelm() unexpected error: %v", err)
				return
			}

			if planType != PlanTypeHelm {
				t.Errorf("tryHelm() type = %v, want %v", planType, PlanTypeHelm)
			}

			if plan == nil {
				t.Error("tryHelm() plan is nil")
			}
		})
	}
}

func TestTryKubernetes(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name:    "valid kubernetes plan",
			jsonStr: `{"k8s_content_diff": [{"kind": "Deployment", "name": "test"}]}`,
			wantErr: false,
		},
		{
			name:    "kubernetes plan with null k8s_content_diff",
			jsonStr: `{"k8s_content_diff": null}`,
			wantErr: true,
		},
		{
			name:    "not a kubernetes plan",
			jsonStr: `{"resource_changes": []}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planType, plan, err := tryKubernetes(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("tryKubernetes() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("tryKubernetes() unexpected error: %v", err)
				return
			}

			if planType != PlanTypeKubernetes {
				t.Errorf("tryKubernetes() type = %v, want %v", planType, PlanTypeKubernetes)
			}

			if plan == nil {
				t.Error("tryKubernetes() plan is nil")
			}
		})
	}
}

func TestMustParseFunctions(t *testing.T) {
	t.Run("MustParseTerraform panics on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseTerraform did not panic on invalid input")
			}
		}()
		MustParseTerraform(`{"invalid": true}`)
	})

	t.Run("MustParseTerraform succeeds on valid", func(t *testing.T) {
		plan := MustParseTerraform(`{"resource_changes": []}`)
		if plan == nil {
			t.Error("MustParseTerraform returned nil")
		}
	})

	t.Run("MustParseHelm panics on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseHelm did not panic on invalid input")
			}
		}()
		MustParseHelm(`{"invalid": true}`)
	})

	t.Run("MustParseHelm succeeds on valid", func(t *testing.T) {
		plan := MustParseHelm(`{"helm_content_diff": []}`)
		if plan == nil {
			t.Error("MustParseHelm returned nil")
		}
	})

	t.Run("MustParseKubernetes panics on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseKubernetes did not panic on invalid input")
			}
		}()
		MustParseKubernetes(`{"invalid": true}`)
	})

	t.Run("MustParseKubernetes succeeds on valid", func(t *testing.T) {
		plan := MustParseKubernetes(`{"k8s_content_diff": []}`)
		if plan == nil {
			t.Error("MustParseKubernetes returned nil")
		}
	})
}
