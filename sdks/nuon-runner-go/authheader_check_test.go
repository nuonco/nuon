package nuonrunner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestPublicEndpointsSendNoAuthHeader(t *testing.T) {
	var mu sync.Mutex
	authByPath := map[string]string{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		authByPath[r.URL.Path] = r.Header.Get("Authorization")
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("[]"))
		default:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("{}"))
		}
	}))
	defer srv.Close()

	c, err := New(WithURL(srv.URL), WithAuthToken("secret-token"), WithRunnerID("rnr_test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	// authenticated call → must carry the bearer token
	if _, err := c.CreateHeartBeat(ctx, &models.ServiceCreateRunnerHeartBeatRequest{}); err != nil {
		t.Fatalf("CreateHeartBeat: %v", err)
	}

	// public calls → must NOT carry an Authorization header
	_, _ = c.GetProcessShutdowns(ctx, "prc_test")
	_, _ = c.RunnerAuthAWS(ctx, &models.ServiceRunnerAuthAWSRequest{})

	mu.Lock()
	defer mu.Unlock()

	hbPath := "/v1/runners/rnr_test/heart-beats"
	if got := authByPath[hbPath]; got != "Bearer secret-token" {
		t.Errorf("authenticated heartbeat: got Authorization %q, want %q", got, "Bearer secret-token")
	}

	shutdownPath := "/v1/runners/rnr_test/processes/prc_test/shutdowns"
	if got := authByPath[shutdownPath]; got != "" {
		t.Errorf("public shutdowns: got Authorization %q, want empty", got)
	}

	authPath := "/v1/runner-auth/aws"
	if got := authByPath[authPath]; got != "" {
		t.Errorf("public runner-auth: got Authorization %q, want empty", got)
	}
}
