package nuon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"0.19.1109","git_ref":"abc"}`))
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	got, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "0.19.1109" || got.GitRef != "abc" {
		t.Fatalf("got %+v", got)
	}
}

func TestGetVersionHonorsContext(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := New(WithURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := client.GetVersion(ctx)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a timeout error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetVersion did not return after the context deadline")
	}
	<-started
}
