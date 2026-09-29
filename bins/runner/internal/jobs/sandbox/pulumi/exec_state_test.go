package pulumi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadPulumiState_SurvivesCancelledContext(t *testing.T) {
	var gotBody []byte
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

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
