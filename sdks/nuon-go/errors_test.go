package nuon

import (
	"testing"
)

func TestHTTPAPIErrorDecodesResponse(t *testing.T) {
	decoded := newHTTPAPIError(
		409,
		`{"error":"runner unavailable","user_error":true,"description":"Wait for the runner and try again."}`,
	)

	userErr, ok := ToUserError(decoded)
	if !ok {
		t.Fatal("expected a user error")
	}
	if userErr.Error != "runner unavailable" {
		t.Fatalf("unexpected error: %q", userErr.Error)
	}
	if userErr.Description != "Wait for the runner and try again." {
		t.Fatalf("unexpected description: %q", userErr.Description)
	}
	if !decoded.(stderrResponse).IsCode(409) {
		t.Fatal("expected status 409")
	}
}

func TestHTTPAPIErrorPreservesNonJSONBody(t *testing.T) {
	input := newHTTPAPIError(502, "upstream unavailable")
	payload := input.(stderrResponse).GetPayload()
	if payload.Description != "HTTP 502" {
		t.Fatalf("unexpected description: %q", payload.Description)
	}
	if payload.Error != "upstream unavailable" {
		t.Fatalf("unexpected error: %q", payload.Error)
	}
}
