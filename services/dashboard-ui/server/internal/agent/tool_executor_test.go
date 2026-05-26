package agent

import (
	"context"
	"testing"
)

func TestExecuteUnknownTool(t *testing.T) {
	e := NewToolExecutor("https://api.example.com")
	_, err := e.Execute(context.Background(), "token", "org", ToolCall{Name: "nonexistent_tool", Args: "{}"})
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if err.Error() != "unknown tool: nonexistent_tool" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExecuteInvalidArgs(t *testing.T) {
	e := NewToolExecutor("https://api.example.com")
	_, err := e.Execute(context.Background(), "token", "org", ToolCall{Name: "list_apps", Args: "not json"})
	if err == nil {
		t.Fatal("expected error for invalid args")
	}
}
