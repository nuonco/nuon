package plandiff

import (
	"strings"
	"testing"
)

func TestFormatPlan(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		wantErr     bool
		errContains string
		contains    []string
	}{
		{
			name:        "empty string",
			jsonStr:     "",
			wantErr:     true,
			errContains: "empty plan",
		},
		{
			name:        "invalid JSON",
			jsonStr:     "{not valid json}",
			wantErr:     true,
			errContains: "unable to detect",
		},
		{
			name:        "unknown plan type",
			jsonStr:     `{"unknown_field": "value"}`,
			wantErr:     true,
			errContains: "unable to detect",
		},
		{
			name: "terraform plan",
			jsonStr: `{
				"resource_changes": [
					{
						"address": "aws_instance.example",
						"type": "aws_instance",
						"name": "example",
						"change": {
							"actions": ["create"]
						}
					}
				]
			}`,
			wantErr:  false,
			contains: []string{"1 to create", "aws_instance"},
		},
		{
			name: "helm plan",
			jsonStr: `{
				"plan": "upgrade",
				"op": "install",
				"helm_content_diff": [
					{
						"kind": "Deployment",
						"name": "my-app",
						"namespace": "default",
						"entries": [
							{"type": 1, "path": "spec.replicas", "applied": "3"}
						]
					}
				]
			}`,
			wantErr:  false,
			contains: []string{"1 to add", "Deployment", "my-app"},
		},
		{
			name: "kubernetes plan",
			jsonStr: `{
				"plan": "apply",
				"op": "create",
				"k8s_content_diff": [
					{
						"kind": "Service",
						"name": "my-svc",
						"namespace": "default",
						"type": 1
					}
				]
			}`,
			wantErr:  false,
			contains: []string{"1 to add", "Service", "my-svc"},
		},
		{
			name: "terraform plan with no changes",
			jsonStr: `{
				"resource_changes": []
			}`,
			wantErr:  false,
			contains: []string{"No changes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPlan(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FormatPlan() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("FormatPlan() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("FormatPlan() unexpected error: %v", err)
				return
			}

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatPlan() output missing %q\nOutput: %s", want, result)
				}
			}
		})
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name       string
		jsonStr    string
		wantChange bool
		wantErr    bool
	}{
		{
			name:    "empty string",
			jsonStr: "",
			wantErr: true,
		},
		{
			name: "terraform no changes",
			jsonStr: `{
				"resource_changes": []
			}`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "terraform with changes",
			jsonStr: `{
				"resource_changes": [
					{
						"address": "aws_instance.example",
						"type": "aws_instance",
						"name": "example",
						"change": {"actions": ["create"]}
					}
				]
			}`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "terraform only no-op",
			jsonStr: `{
				"resource_changes": [
					{
						"address": "aws_instance.example",
						"type": "aws_instance",
						"name": "example",
						"change": {"actions": ["no-op"]}
					}
				]
			}`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "helm no changes",
			jsonStr: `{
				"helm_content_diff": []
			}`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "helm with changes",
			jsonStr: `{
				"helm_content_diff": [
					{"kind": "Deployment", "name": "app", "after": "replicas: 1"}
				]
			}`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "kubernetes no changes",
			jsonStr: `{
				"k8s_content_diff": []
			}`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "kubernetes with changes",
			jsonStr: `{
				"k8s_content_diff": [
					{"kind": "Service", "name": "svc", "type": 1}
				]
			}`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "kubernetes with errors",
			jsonStr: `{
				"k8s_content_diff": [
					{"kind": "Deployment", "name": "broken", "error": "failed"}
				]
			}`,
			wantChange: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasChanges(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("HasChanges() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("HasChanges() unexpected error: %v", err)
				return
			}

			if got != tt.wantChange {
				t.Errorf("HasChanges() = %v, want %v", got, tt.wantChange)
			}
		})
	}
}

func TestGetPlanType(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		wantType PlanType
		wantErr  bool
	}{
		{
			name:    "empty string",
			jsonStr: "",
			wantErr: true,
		},
		{
			name:     "terraform plan",
			jsonStr:  `{"resource_changes": []}`,
			wantType: PlanTypeTerraform,
			wantErr:  false,
		},
		{
			name:     "helm plan",
			jsonStr:  `{"helm_content_diff": []}`,
			wantType: PlanTypeHelm,
			wantErr:  false,
		},
		{
			name:     "kubernetes plan",
			jsonStr:  `{"k8s_content_diff": []}`,
			wantType: PlanTypeKubernetes,
			wantErr:  false,
		},
		{
			name:    "unknown plan",
			jsonStr: `{"other_field": true}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPlanType(tt.jsonStr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPlanType() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPlanType() unexpected error: %v", err)
				return
			}

			if got != tt.wantType {
				t.Errorf("GetPlanType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestFormatPlanCastErrors(t *testing.T) {
	// These tests verify that FormatPlan properly handles the internal type casts
	// by testing with real plan detection scenarios

	t.Run("terraform plan cast", func(t *testing.T) {
		jsonStr := `{"resource_changes": [{"address": "test", "type": "test", "name": "test", "change": {"actions": ["create"]}}]}`
		_, err := FormatPlan(jsonStr)
		if err != nil {
			t.Errorf("FormatPlan() unexpected cast error: %v", err)
		}
	})

	t.Run("helm plan cast", func(t *testing.T) {
		jsonStr := `{"helm_content_diff": [{"kind": "Deployment", "name": "test"}]}`
		_, err := FormatPlan(jsonStr)
		if err != nil {
			t.Errorf("FormatPlan() unexpected cast error: %v", err)
		}
	})

	t.Run("kubernetes plan cast", func(t *testing.T) {
		jsonStr := `{"k8s_content_diff": [{"kind": "Service", "name": "test", "type": 1}]}`
		_, err := FormatPlan(jsonStr)
		if err != nil {
			t.Errorf("FormatPlan() unexpected cast error: %v", err)
		}
	})
}

func TestHasChangesCastErrors(t *testing.T) {
	// These tests verify that HasChanges properly handles the internal type casts

	t.Run("terraform plan cast", func(t *testing.T) {
		jsonStr := `{"resource_changes": [{"address": "test", "type": "test", "name": "test", "change": {"actions": ["create"]}}]}`
		_, err := HasChanges(jsonStr)
		if err != nil {
			t.Errorf("HasChanges() unexpected cast error: %v", err)
		}
	})

	t.Run("helm plan cast", func(t *testing.T) {
		jsonStr := `{"helm_content_diff": [{"kind": "Deployment", "name": "test", "after": "data"}]}`
		_, err := HasChanges(jsonStr)
		if err != nil {
			t.Errorf("HasChanges() unexpected cast error: %v", err)
		}
	})

	t.Run("kubernetes plan cast", func(t *testing.T) {
		jsonStr := `{"k8s_content_diff": [{"kind": "Service", "name": "test", "type": 1}]}`
		_, err := HasChanges(jsonStr)
		if err != nil {
			t.Errorf("HasChanges() unexpected cast error: %v", err)
		}
	})
}

func TestComplexTerraformPlan(t *testing.T) {
	jsonStr := `{
		"resource_drift": [
			{
				"address": "aws_instance.drifted",
				"type": "aws_instance",
				"name": "drifted",
				"change": {
					"actions": ["update"],
					"before": {"tags": {"Name": "old"}},
					"after": {"tags": {"Name": "new"}}
				}
			}
		],
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"instance_type": "t3.micro"},
					"after_unknown": {"id": true, "public_ip": true}
				}
			},
			{
				"address": "aws_instance.old",
				"type": "aws_instance",
				"name": "old",
				"change": {
					"actions": ["delete"],
					"before": {"instance_type": "t2.micro"},
					"after": null
				}
			},
			{
				"address": "aws_instance.replace",
				"type": "aws_instance",
				"name": "replace",
				"change": {
					"actions": ["replace"],
					"before": {"ami": "ami-old"},
					"after": {"ami": "ami-new"}
				}
			}
		],
		"output_changes": {
			"web_ip": {
				"actions": ["create"],
				"before": null,
				"after": null,
				"after_unknown": true
			}
		}
	}`

	result, err := FormatPlan(jsonStr)
	if err != nil {
		t.Fatalf("FormatPlan() error: %v", err)
	}

	expectations := []string{
		"to create",
		"to delete",
		"to replace",
		"Resource Drift",
		"Resource Changes",
		"Output Changes",
		"aws_instance.web",
		"aws_instance.old",
		"aws_instance.replace",
		"web_ip",
	}

	for _, want := range expectations {
		if !strings.Contains(result, want) {
			t.Errorf("FormatPlan() output missing %q", want)
		}
	}

	hasChanges, err := HasChanges(jsonStr)
	if err != nil {
		t.Fatalf("HasChanges() error: %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() = false, want true for complex plan with changes")
	}
}

func TestComplexHelmPlan(t *testing.T) {
	jsonStr := `{
		"plan": "upgrade my-release",
		"op": "upgrade",
		"helm_content_diff": [
			{
				"api": "apps/v1",
				"kind": "Deployment",
				"name": "frontend",
				"namespace": "production",
				"entries": [
					{"type": 3, "path": "spec.replicas", "original": "2", "applied": "5"},
					{"type": 3, "path": "spec.template.spec.containers[0].image", "original": "app:v1", "applied": "app:v2"}
				]
			},
			{
				"api": "v1",
				"kind": "ConfigMap",
				"name": "app-config",
				"namespace": "production",
				"entries": [
					{"type": 1, "path": "data.NEW_KEY", "applied": "new_value"}
				]
			},
			{
				"kind": "Secret",
				"name": "old-secret",
				"namespace": "production",
				"before": "some-data"
			}
		]
	}`

	result, err := FormatPlan(jsonStr)
	if err != nil {
		t.Fatalf("FormatPlan() error: %v", err)
	}

	expectations := []string{
		"to change",
		"to add",
		"to destroy",
		"Deployment",
		"frontend",
		"ConfigMap",
		"app-config",
		"Secret",
		"old-secret",
		"production",
	}

	for _, want := range expectations {
		if !strings.Contains(result, want) {
			t.Errorf("FormatPlan() output missing %q", want)
		}
	}
}

func TestComplexKubernetesPlan(t *testing.T) {
	jsonStr := `{
		"plan": "apply namespace/manifests",
		"op": "apply",
		"k8s_content_diff": [
			{
				"_version": "1",
				"api": "apps/v1",
				"kind": "Deployment",
				"name": "api-server",
				"namespace": "backend",
				"type": 1,
				"dry_run": true
			},
			{
				"_version": "1",
				"api": "v1",
				"kind": "Service",
				"name": "api-service",
				"namespace": "backend",
				"type": 3,
				"entries": [
					{"type": 3, "path": "spec.ports[0].port", "original": "80", "applied": "8080"}
				]
			},
			{
				"_version": "1",
				"kind": "ConfigMap",
				"name": "broken-config",
				"namespace": "backend",
				"error": "validation error: missing required field 'data'"
			}
		]
	}`

	result, err := FormatPlan(jsonStr)
	if err != nil {
		t.Fatalf("FormatPlan() error: %v", err)
	}

	expectations := []string{
		"Plan:",
		"apply namespace/manifests",
		"to add",
		"to change",
		"Deployment",
		"api-server",
		"Service",
		"api-service",
		"Errors",
		"broken-config",
		"validation error",
	}

	for _, want := range expectations {
		if !strings.Contains(result, want) {
			t.Errorf("FormatPlan() output missing %q", want)
		}
	}

	hasChanges, err := HasChanges(jsonStr)
	if err != nil {
		t.Fatalf("HasChanges() error: %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() = false, want true (plan has errors)")
	}
}
