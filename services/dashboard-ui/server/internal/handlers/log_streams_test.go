package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type logStreamTestClient struct {
	nuon.Client
	tailResponses   []*models.ServiceLogStreamTailLogsResponse
	legacyLogs      []*models.AppOtelLogRecord
	legacyNext      string
	tailResponseIdx int
	lastFilters     *nuon.LogStreamLogFilters
}

func (c *logStreamTestClient) LogStreamTailLogs(_ context.Context, _ string, _ string, _ string, filters *nuon.LogStreamLogFilters) (*models.ServiceLogStreamTailLogsResponse, error) {
	c.lastFilters = filters
	response := c.tailResponses[c.tailResponseIdx]
	c.tailResponseIdx++
	return response, nil
}

func (c *logStreamTestClient) LogStreamReadLogsWithNextOffset(_ context.Context, _ string, _ string, _ string, filters *nuon.LogStreamLogFilters) ([]*models.AppOtelLogRecord, string, error) {
	c.lastFilters = filters
	return c.legacyLogs, c.legacyNext, nil
}

func TestParseLogFiltersFromQuery(t *testing.T) {
	tests := map[string]struct {
		query string
		want  *nuon.LogStreamLogFilters
		isNil bool
	}{
		"no params returns nil":              {query: "", isNil: true},
		"only non-filter params returns nil": {query: "wait=1s", isNil: true},
		"single value params map by name": {
			query: "trace_id=tid&q=boom&tool=helm&runner_job_id=job-1",
			want: &nuon.LogStreamLogFilters{
				TraceID:      "tid",
				BodyContains: "boom",
				Tools:        []string{"helm"},
				RunnerJobID:  "job-1",
			},
		},
		"repeatable params keep all values": {
			query: "severity_text=Error&severity_text=Warn&scope_name=oteljob&attr=nuon.tool:helm",
			want: &nuon.LogStreamLogFilters{
				SeverityTexts: []string{"Error", "Warn"},
				ScopeNames:    []string{"oteljob"},
				Attrs:         []string{"nuon.tool:helm"},
			},
		},
		"numeric params parse and invalid ones are dropped": {
			query: "severity_number_min=2&severity_number_max=bad&trace_flags=xyz",
			want: &nuon.LogStreamLogFilters{
				SeverityNumberMin: 2,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			got := parseLogFiltersFromQuery(c)
			if tt.isNil {
				if got != nil {
					t.Fatalf("got %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("got nil filters, want populated")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStreamSessionForwardsFiltersToTail(t *testing.T) {
	client := &logStreamTestClient{
		tailResponses: []*models.ServiceLogStreamTailLogsResponse{
			{Logs: testRunnerJobLogs(), Next: "next"},
			{},
		},
	}

	runFilteredLogStreamSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runTail(ctx)
	})

	if client.lastFilters == nil || client.lastFilters.RunnerJobID != "job-target" {
		t.Errorf("tail request filters = %+v, want runner_job_id=job-target", client.lastFilters)
	}
}

func runFilteredLogStreamSession(t *testing.T, client nuon.Client, run func(context.Context, *streamSession)) []string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []string
	session := &streamSession{
		client:  client,
		l:       zap.NewNop(),
		filters: &nuon.LogStreamLogFilters{RunnerJobID: "job-target"},
		isOpen:  false,
		sendEvent: func(logs []*models.AppOtelLogRecord) {
			for _, log := range logs {
				got = append(got, log.ID)
			}
		},
		sendStatus: func(status string) {
			if status == "complete" {
				cancel()
			}
		},
		sendError: func(string) {},
	}

	run(ctx, session)
	return got
}

func testRunnerJobLogs() []*models.AppOtelLogRecord {
	return []*models.AppOtelLogRecord{
		{ID: "matching", RunnerJobID: "job-target"},
		{ID: "sibling", RunnerJobID: "job-sibling"},
		{ID: "unscoped"},
	}
}

type logStreamErrorClient struct {
	nuon.Client
	tailErr   error
	legacyErr error
	tailResp  *models.ServiceLogStreamTailLogsResponse

	tailCalls   int
	legacyCalls int
}

func (c *logStreamErrorClient) LogStreamTailLogs(_ context.Context, _ string, _ string, _ string, _ *nuon.LogStreamLogFilters) (*models.ServiceLogStreamTailLogsResponse, error) {
	c.tailCalls++
	if c.tailErr != nil {
		return nil, c.tailErr
	}
	return c.tailResp, nil
}

func (c *logStreamErrorClient) LogStreamReadLogsWithNextOffset(_ context.Context, _ string, _ string, _ string, _ *nuon.LogStreamLogFilters) ([]*models.AppOtelLogRecord, string, error) {
	c.legacyCalls++
	if c.legacyErr != nil {
		return nil, "", c.legacyErr
	}
	return nil, "", nil
}

// runErrorSession drives a session whose client errors and records what the
// session sent. When cancelOnFirstError is set (transient-error case), the
// context is cancelled as soon as the first error event is sent, cutting the
// in-flight errorRetryDelay sleep short.
func runErrorSession(t *testing.T, client *logStreamErrorClient, run func(context.Context, *streamSession), cancelOnFirstError bool) (errs []string, elapsed time.Duration) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sent []string
	session := &streamSession{
		client:      client,
		l:           zap.NewNop(),
		logStreamID: "ls-1",
		filters:     &nuon.LogStreamLogFilters{},
		isOpen:      false,
		sendEvent:   func([]*models.AppOtelLogRecord) {},
		sendStatus:  func(string) {},
		sendError: func(msg string) {
			sent = append(sent, msg)
			if cancelOnFirstError {
				cancel()
			}
		},
	}

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer close(done)
		run(ctx, session)
	}()

	select {
	case <-done:
	case <-time.After(3 * errorRetryDelay):
		t.Fatal("session did not return; expected terminal error or cancelled retry")
	}
	cancel()
	<-done

	return sent, time.Since(start)
}

func TestRunTailTerminalErrorDoesNotRetry(t *testing.T) {
	client := &logStreamErrorClient{
		tailErr: &operations.LogStreamTailLogsBadRequest{
			Payload: &models.StderrErrResponse{
				Error:       "invalid request",
				Description: "invalid scope_attr: invalid key \"bad-key\"",
			},
		},
	}

	errs, elapsed := runErrorSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runTail(ctx)
	}, false)

	if client.tailCalls != 1 {
		t.Errorf("tail calls = %d, want 1 (no retry on 4xx)", client.tailCalls)
	}
	if len(errs) != 1 || errs[0] != `invalid scope_attr: invalid key "bad-key"` {
		t.Errorf("sent errors = %v, want one event with the ctl-api description", errs)
	}
	if elapsed >= errorRetryDelay {
		t.Errorf("session took %s, want return before the retry delay", elapsed)
	}
}

func TestRunLegacyTerminalErrorDoesNotRetry(t *testing.T) {
	client := &logStreamErrorClient{
		legacyErr: &operations.LogStreamReadLogsBadRequest{
			Payload: &models.StderrErrResponse{Description: "invalid request input"},
		},
	}

	errs, elapsed := runErrorSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runLegacy(ctx, "")
	}, false)

	if client.legacyCalls != 1 {
		t.Errorf("legacy calls = %d, want 1 (no retry on 4xx)", client.legacyCalls)
	}
	if len(errs) != 1 || errs[0] != "invalid request input" {
		t.Errorf("sent errors = %v, want one event with the ctl-api description", errs)
	}
	if elapsed >= errorRetryDelay {
		t.Errorf("session took %s, want return before the retry delay", elapsed)
	}
}

func TestRunTailTransientErrorRetries(t *testing.T) {
	client := &logStreamErrorClient{tailErr: io.EOF}

	errs, _ := runErrorSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runTail(ctx)
	}, true)

	if client.tailCalls != 1 {
		t.Errorf("tail calls = %d, want 1 before the retry delay fires", client.tailCalls)
	}
	if len(errs) != 1 || errs[0] != "Polling failed" {
		t.Errorf("sent errors = %v, want one \"Polling failed\" event", errs)
	}
}
