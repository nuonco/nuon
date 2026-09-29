package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
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

func TestCreateCloudConnectionErrorIncludesDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/cloud-connections" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid role","user_error":true,"description":"principal must be an IAM role ARN in account 133456789012"}`)
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL), WithOrgID("org-example"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateCloudConnection(context.Background(), &models.ServiceCreateRequest{
		Name: "acme-production", Platform: "aws", TargetID: "133456789012",
		Principal: "arn:aws:iam::123456789012:role/nuon-cloud-connection", Preset: models.AppCloudConnectionPresetStacks,
	})
	userErr, ok := ToUserError(err)
	if !ok {
		t.Fatalf("expected decoded user error, got %v", err)
	}
	if userErr.Description != "principal must be an IAM role ARN in account 133456789012" {
		t.Fatalf("unexpected description: %q", userErr.Description)
	}
}
