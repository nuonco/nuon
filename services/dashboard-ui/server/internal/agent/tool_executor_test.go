package agent

import (
	"testing"
)

func TestResolveToolCall(t *testing.T) {
	e := NewToolExecutor("https://api.example.com")

	tests := []struct {
		name       string
		args       map[string]any
		wantMethod string
		wantPath   string
	}{
		{
			name:       "list_apps",
			args:       map[string]any{},
			wantMethod: "GET",
			wantPath:   "/v1/apps",
		},
		{
			name:       "get_app",
			args:       map[string]any{"app_id": "app123"},
			wantMethod: "GET",
			wantPath:   "/v1/apps/app123",
		},
		{
			name:       "create_install",
			args:       map[string]any{"app_id": "app123", "name": "test", "region": "us-west-2"},
			wantMethod: "POST",
			wantPath:   "/v1/apps/app123/installs",
		},
		{
			name:       "get_cloud_regions",
			args:       map[string]any{"platform": "aws"},
			wantMethod: "GET",
			wantPath:   "/v1/general/cloud-platform/aws/regions",
		},
		{
			name:       "create_terraform_config",
			args:       map[string]any{"app_id": "app1", "component_id": "comp1"},
			wantMethod: "POST",
			wantPath:   "/v1/apps/app1/components/comp1/configs/terraform-module",
		},
		{
			name:       "list_vcs_connections",
			args:       map[string]any{},
			wantMethod: "GET",
			wantPath:   "/v1/vcs/connections",
		},
		{
			name:       "get_step_logs",
			args:       map[string]any{"log_stream_id": "ls123"},
			wantMethod: "GET",
			wantPath:   "/v1/log-streams/ls123/logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, path, _, err := e.resolveToolCall(tt.name, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if method != tt.wantMethod {
				t.Errorf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Errorf("path = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

func TestResolveToolCallUnknown(t *testing.T) {
	e := NewToolExecutor("https://api.example.com")
	_, _, _, err := e.resolveToolCall("nonexistent_tool", map[string]any{})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
