package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type sseScriptStep struct {
	res sseFetchResult
	err error
}

// runScriptedStream runs runSSEStream against a scripted sequence of fetch
// results, cancelling the request context once the script is exhausted, and
// returns the raw response body.
func runScriptedStream(t *testing.T, steps []sseScriptStep, mod func(*sseStreamConfig)) string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Request = httptest.NewRequest(http.MethodGet, "/sse", nil).WithContext(ctx)

	i := 0
	cfg := sseStreamConfig{
		ClientErrMsg:    "fetch failed",
		PollInterval:    time.Millisecond,
		ErrorRetryDelay: time.Millisecond,
		Log:             zap.NewNop(),
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			if i >= len(steps) {
				cancel()
				return sseFetchResult{}, fmt.Errorf("script exhausted: %w", errSSESilentRetry)
			}
			step := steps[i]
			i++
			return step.res, step.err
		},
	}
	if mod != nil {
		mod(&cfg)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runSSEStream(c, cfg)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("runSSEStream did not return")
	}
	return w.Body.String()
}

func event(name, data string) sseEvent {
	return sseEvent{Name: name, Data: []byte(data)}
}

func TestRunSSEStreamHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Request = httptest.NewRequest(http.MethodGet, "/sse", nil).WithContext(ctx)

	runSSEStream(c, sseStreamConfig{
		Fetch: func(ctx context.Context) (sseFetchResult, error) {
			return sseFetchResult{}, nil
		},
	})

	if got := w.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
	if got := w.Header().Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q, want no", got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRunSSEStreamHashDedupe(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{res: sseFetchResult{Events: []sseEvent{event("build", `{"v":1}`)}}},
		{res: sseFetchResult{Events: []sseEvent{event("build", `{"v":1}`)}}},
		{res: sseFetchResult{Events: []sseEvent{event("build", `{"v":2}`)}}},
	}, nil)

	if got := strings.Count(body, "event: build\n"); got != 2 {
		t.Errorf("build events = %d, want 2 (deduped)\nbody:\n%s", got, body)
	}
	if !strings.Contains(body, `data: {"v":1}`) || !strings.Contains(body, `data: {"v":2}`) {
		t.Errorf("missing payloads\nbody:\n%s", body)
	}
}

func TestRunSSEStreamPerEventHashes(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{res: sseFetchResult{Events: []sseEvent{event("deploy", `{"v":1}`), event("workflow", `{"w":1}`)}}},
		// workflow omitted: its hash must survive, deploy unchanged
		{res: sseFetchResult{Events: []sseEvent{event("deploy", `{"v":1}`)}}},
		// workflow returns unchanged: must not re-emit
		{res: sseFetchResult{Events: []sseEvent{event("deploy", `{"v":1}`), event("workflow", `{"w":1}`)}}},
	}, nil)

	if got := strings.Count(body, "event: workflow\n"); got != 1 {
		t.Errorf("workflow events = %d, want 1\nbody:\n%s", got, body)
	}
}

func TestRunSSEStreamPartialEmitOnError(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{
			res: sseFetchResult{Events: []sseEvent{event("active-workflows", `[1]`)}},
			err: fmt.Errorf("history fetch failed"),
		},
	}, nil)

	if !strings.Contains(body, "event: active-workflows\ndata: [1]") {
		t.Errorf("partial events not emitted before error\nbody:\n%s", body)
	}
	if !strings.Contains(body, `event: fetch-error`) || !strings.Contains(body, `{"error":"fetch failed"}`) {
		t.Errorf("missing fetch-error event\nbody:\n%s", body)
	}
	if idx := strings.Index(body, "event: active-workflows"); idx > strings.Index(body, "event: fetch-error") {
		t.Errorf("partial events must emit before fetch-error\nbody:\n%s", body)
	}
}

func TestRunSSEStreamSilentRetry(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{err: fmt.Errorf("marshal build: %w", errSSESilentRetry)},
		{res: sseFetchResult{Events: []sseEvent{event("build", `{"v":1}`)}}},
	}, nil)

	if strings.Contains(body, "fetch-error") {
		t.Errorf("silent retry must not emit fetch-error\nbody:\n%s", body)
	}
	if !strings.Contains(body, "event: build\n") {
		t.Errorf("expected recovery after silent retry\nbody:\n%s", body)
	}
}

func TestRunSSEStreamFinishedGracePeriod(t *testing.T) {
	finished := sseFetchResult{
		Events:   []sseEvent{event("build", `{"status":"success"}`)},
		Finished: true,
	}
	// Script never exhausts within the grace period; the stream must close
	// itself. 50 steps at 1ms polls comfortably outlast a 10ms grace.
	steps := make([]sseScriptStep, 50)
	for i := range steps {
		steps[i] = sseScriptStep{res: finished}
	}

	body := runScriptedStream(t, steps, func(cfg *sseStreamConfig) {
		cfg.FinishedPollInterval = time.Millisecond
		cfg.FinishedGracePeriod = 10 * time.Millisecond
	})

	if got := strings.Count(body, "event: finished\ndata: true"); got != 1 {
		t.Errorf("finished events = %d, want 1\nbody:\n%s", got, body)
	}
}

func TestRunSSEStreamNoFinishedWithoutGracePeriod(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{res: sseFetchResult{Events: []sseEvent{event("builds", `[1]`)}, Finished: true}},
	}, nil)

	if strings.Contains(body, "event: finished") {
		t.Errorf("timeline streams (no grace period) must not emit finished\nbody:\n%s", body)
	}
}

func TestRunSSEStreamUnnamedEvent(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{res: sseFetchResult{Events: []sseEvent{event("", `{"v":1}`)}}},
	}, nil)

	if !strings.Contains(body, "data: {\"v\":1}\n\n") {
		t.Errorf("missing unnamed event payload\nbody:\n%s", body)
	}
	if strings.Contains(body, "event: \n") {
		t.Errorf("unnamed event must not write an event: line\nbody:\n%s", body)
	}
}

func TestRunSSEStreamKeepalive(t *testing.T) {
	body := runScriptedStream(t, []sseScriptStep{
		{res: sseFetchResult{Events: []sseEvent{event("build", `{"v":1}`)}}},
	}, nil)

	if !strings.Contains(body, ": keepalive\n\n") {
		t.Errorf("missing keepalive comment\nbody:\n%s", body)
	}
}
