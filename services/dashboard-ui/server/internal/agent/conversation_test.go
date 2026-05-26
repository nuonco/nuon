package agent

import (
	"testing"
)

func TestConversationStore(t *testing.T) {
	store := NewConversationStore()

	conv := store.Create("conv-1", "org-1")
	if conv.ID != "conv-1" || conv.OrgID != "org-1" {
		t.Fatal("unexpected conversation")
	}

	got := store.Get("conv-1")
	if got != conv {
		t.Fatal("expected same conversation")
	}

	if store.Get("nonexistent") != nil {
		t.Fatal("expected nil for nonexistent conversation")
	}

	store.Create("conv-2", "org-1")
	store.Create("conv-3", "org-2")

	list := store.ListByOrg("org-1")
	if len(list) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(list))
	}

	list2 := store.ListByOrg("org-2")
	if len(list2) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(list2))
	}

	store.Delete("conv-1")
	if store.Get("conv-1") != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestConversationAppendMessage(t *testing.T) {
	store := NewConversationStore()
	conv := store.Create("conv-1", "org-1")

	conv.AppendMessage(Message{Role: "user", Content: "hello"})
	conv.AppendMessage(Message{Role: "assistant", Content: "hi there"})

	if len(conv.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(conv.Messages))
	}

	if conv.Messages[0].Content != "hello" {
		t.Fatalf("expected first message to be 'hello', got '%s'", conv.Messages[0].Content)
	}
}

func TestConversationMessageCap(t *testing.T) {
	store := NewConversationStore()
	conv := store.Create("conv-1", "org-1")

	for i := 0; i < maxMessages+20; i++ {
		conv.AppendMessage(Message{Role: "user", Content: "msg"})
	}

	if len(conv.Messages) != maxMessages {
		t.Fatalf("expected %d messages, got %d", maxMessages, len(conv.Messages))
	}
}
