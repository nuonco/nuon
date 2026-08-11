package handlers

import (
	"context"
	"testing"

	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type logStreamTestClient struct {
	nuon.Client
	tailResponses   []*models.ServiceLogStreamTailLogsResponse
	legacyLogs      []*models.AppOtelLogRecord
	legacyNext      string
	tailResponseIdx int
}

func (c *logStreamTestClient) LogStreamTailLogs(context.Context, string, string, string) (*models.ServiceLogStreamTailLogsResponse, error) {
	response := c.tailResponses[c.tailResponseIdx]
	c.tailResponseIdx++
	return response, nil
}

func (c *logStreamTestClient) LogStreamReadLogsWithNextOffset(context.Context, string, string, string) ([]*models.AppOtelLogRecord, string, error) {
	return c.legacyLogs, c.legacyNext, nil
}

func TestFilterLogsByRunnerJobID(t *testing.T) {
	logs := []*models.AppOtelLogRecord{
		{ID: "matching", RunnerJobID: "job-target"},
		{ID: "sibling", RunnerJobID: "job-sibling"},
		{ID: "unscoped"},
	}

	tests := map[string]struct {
		runnerJobID string
		wantIDs     []string
	}{
		"empty filter preserves every log": {
			wantIDs: []string{"matching", "sibling", "unscoped"},
		},
		"job filter keeps only matching logs": {
			runnerJobID: "job-target",
			wantIDs:     []string{"matching"},
		},
		"unknown job returns no logs": {
			runnerJobID: "job-unknown",
			wantIDs:     []string{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := filterLogsByRunnerJobID(logs, tt.runnerJobID)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("got %d logs, want %d", len(got), len(tt.wantIDs))
			}
			for i, wantID := range tt.wantIDs {
				if got[i].ID != wantID {
					t.Errorf("log %d ID = %q, want %q", i, got[i].ID, wantID)
				}
			}
		})
	}
}

func TestStreamSessionFiltersTailLogsByRunnerJobID(t *testing.T) {
	client := &logStreamTestClient{
		tailResponses: []*models.ServiceLogStreamTailLogsResponse{
			{Logs: testRunnerJobLogs(), Next: "next"},
			{},
		},
	}

	got := runFilteredLogStreamSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runTail(ctx)
	})
	assertLogIDs(t, got, []string{"matching"})
}

func TestStreamSessionFiltersLegacyLogsByRunnerJobID(t *testing.T) {
	client := &logStreamTestClient{legacyLogs: testRunnerJobLogs()}

	got := runFilteredLogStreamSession(t, client, func(ctx context.Context, session *streamSession) {
		session.runLegacy(ctx, "")
	})
	assertLogIDs(t, got, []string{"matching"})
}

func runFilteredLogStreamSession(t *testing.T, client nuon.Client, run func(context.Context, *streamSession)) []string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []string
	session := &streamSession{
		client:      client,
		l:           zap.NewNop(),
		runnerJobID: "job-target",
		isOpen:      false,
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

func assertLogIDs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got log IDs %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("log %d ID = %q, want %q", i, got[i], want[i])
		}
	}
}
