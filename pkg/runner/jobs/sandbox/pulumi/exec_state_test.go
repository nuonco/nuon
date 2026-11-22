package pulumi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUploadPulumiState_SurvivesCancelledContext proves the fix: a job context
// that is already cancelled (mid-deploy cancel) must NOT prevent state from
// being persisted. Before the fix the upload used the cancelled context and the
// POST failed, so created resources were lost and the retry recreated them.
func TestUploadPulumiState_SurvivesCancelledContext(t *testing.T) {
	var gotBody []byte
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Simulate the job being cancelled before we try to persist state.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	state := []byte(`{"resources":[{"urn":"forgejo"}]}`)
	if err := uploadPulumiState(ctx, srv.URL, "tok", "ws123", "job456", state); err != nil {
		t.Fatalf("upload should succeed despite cancelled parent ctx, got: %v", err)
	}

	if string(gotBody) != string(state) {
		t.Fatalf("server received %q, want %q", gotBody, state)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("missing/incorrect auth header: %q", gotAuth)
	}
}
