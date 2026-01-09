package service

import (
	"regexp"
	"testing"
)

func TestGenerateStateNonce(t *testing.T) {
	// Generate a state nonce
	state, err := generateStateNonce()
	if err != nil {
		t.Fatalf("generateStateNonce() returned error: %v", err)
	}

	// Test length - should be reasonable length after base64 encoding and stripping
	// 32 bytes -> ~43 base64 chars, after stripping non-alphanumeric should still be substantial
	minLength := 20
	if len(state) < minLength {
		t.Errorf("generateStateNonce() returned state with length %d, want at least %d", len(state), minLength)
	}

	// Test that there are no non-alphanumeric characters
	nonAlphaNum := regexp.MustCompile("[^a-zA-Z0-9]")
	if nonAlphaNum.MatchString(state) {
		t.Errorf("generateStateNonce() returned state with non-alphanumeric characters: %q", state)
	}

	// Test uniqueness - generate another and ensure they're different
	state2, err := generateStateNonce()
	if err != nil {
		t.Fatalf("generateStateNonce() second call returned error: %v", err)
	}
	if state == state2 {
		t.Error("generateStateNonce() returned identical states on consecutive calls")
	}
}
