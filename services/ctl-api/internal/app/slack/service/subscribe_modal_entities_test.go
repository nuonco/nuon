package service

import (
	"strings"
	"testing"
)

func TestBuildSuggestionOptions(t *testing.T) {
	t.Run("empty input still returns an options key", func(t *testing.T) {
		out := buildSuggestionOptions(nil)
		opts, ok := out["options"].([]any)
		if !ok {
			t.Fatalf("expected options to be []any, got %T", out["options"])
		}
		if len(opts) != 0 {
			t.Fatalf("expected zero options for nil input, got %d", len(opts))
		}
	})

	t.Run("each item maps to a Slack option with bare name as text", func(t *testing.T) {
		out := buildSuggestionOptions([]suggestionItem{
			{ID: "cmpid1", Name: "frontend"},
			{ID: "cmpid2", Name: "backend"},
		})
		opts := out["options"].([]any)
		if len(opts) != 2 {
			t.Fatalf("expected 2 options, got %d", len(opts))
		}
		first := opts[0].(map[string]any)
		if first["value"] != "cmpid1" {
			t.Fatalf("expected value=cmpid1, got %v", first["value"])
		}
		txt := first["text"].(map[string]any)
		if txt["type"] != "plain_text" {
			t.Fatalf("expected type=plain_text, got %v", txt["type"])
		}
		if txt["text"] != "frontend" {
			t.Fatalf("expected bare name 'frontend', got %v", txt["text"])
		}
	})

	t.Run("name longer than 75 bytes is truncated", func(t *testing.T) {
		long := strings.Repeat("x", 200)
		out := buildSuggestionOptions([]suggestionItem{{ID: "id", Name: long}})
		opts := out["options"].([]any)
		txt := opts[0].(map[string]any)["text"].(map[string]any)
		got := txt["text"].(string)
		if len(got) > 75 {
			t.Fatalf("expected truncated label <=75 bytes, got %d", len(got))
		}
	})
}

func TestSubscribeEntitiesActionIDsAreDistinct(t *testing.T) {
	if subscribeEntitiesActionIDComponents == subscribeEntitiesActionIDActions {
		t.Fatal("components and actions action_ids must differ")
	}
	if subscribeEntitiesActionIDComponents == subscribeEntitiesActionIDInstalls {
		t.Fatal("components and installs action_ids must differ")
	}
	if subscribeEntitiesActionIDActions == subscribeEntitiesActionIDInstalls {
		t.Fatal("actions and installs action_ids must differ")
	}
}
