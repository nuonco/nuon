package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

func writeJobLog(t *testing.T, dir, jobID string, lines ...string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(statestore.JobLogKey(jobID)))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600))
}

func writeOTELJobLogs(t *testing.T, dir, name string, records ...string) {
	t.Helper()
	path := filepath.Join(dir, statestore.JobLogsPrefix, name+otelLogsSuffix)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	batch := fmt.Sprintf(`{"resourceLogs":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"runner"}}]},"scopeLogs":[{"scope":{"name":"oteljob"},"logRecords":[%s]}]}]}`, strings.Join(records, ","))
	require.NoError(t, os.WriteFile(path, []byte(batch+"\n"), 0o600))
}

func otelJobLogRecord(jobID, timestamp, level, message string) string {
	return fmt.Sprintf(`{"timeUnixNano":"%s","severityText":%q,"body":{"stringValue":%q},"attributes":[{"key":"runner_job.id","value":{"stringValue":%q}},{"key":"component","value":{"stringValue":"api"}}]}`, timestamp, level, message, jobID)
}

func TestPortalLogsList(t *testing.T) {
	p, dir := testPortal(t)
	now := time.Now().UTC().Truncate(time.Second)
	earlier := now.Add(-time.Hour)

	writeStateObject(t, dir, "status.json", statestore.Status{
		InstallID: "inst-1",
		Steps: []statestore.StepStatus{
			{ID: "job-install-1", Name: "deploy api", Status: "finished", StartedAt: &earlier},
			{ID: "job-install-missing", Name: "sync api", Status: "finished", StartedAt: &earlier},
		},
	})
	writeStateObject(t, dir, operation.RunStatusKey("run-2"), operation.RunStatus{
		RunID:     "run-2",
		RefID:     "restart-api",
		RefName:   "Restart API",
		StartedAt: now,
		Steps: []operation.RunStep{
			{ID: "step-1", Name: "restart", Kind: "action", JobID: "run-2-step-1", Status: "finished"},
		},
	})
	writeJobLog(t, dir, "job-install-1", `{"level":"info","ts":1700000000,"msg":"install"}`)
	writeJobLog(t, dir, "run-2-step-1", `{"level":"info","ts":1700000001,"msg":"restart"}`)
	writeJobLog(t, dir, "job-orphan", `{"level":"info","ts":1700000002,"msg":"orphan"}`)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Jobs []jobLogSummary `json:"jobs"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Len(t, payload.Jobs, 4)

	require.Equal(t, "run-2-step-1", payload.Jobs[0].JobID)
	require.Equal(t, logSourceOperation, payload.Jobs[0].Source)
	require.Equal(t, "Restart API", payload.Jobs[0].RefName)
	require.Equal(t, "run-2", payload.Jobs[0].RunID)
	require.True(t, payload.Jobs[0].LogsAvailable)

	require.Equal(t, "job-install-1", payload.Jobs[1].JobID)
	require.Equal(t, logSourceInstall, payload.Jobs[1].Source)
	require.Equal(t, "deploy api", payload.Jobs[1].Name)
	require.True(t, payload.Jobs[1].LogsAvailable)

	require.Equal(t, "job-install-missing", payload.Jobs[2].JobID)
	require.Equal(t, logSourceInstall, payload.Jobs[2].Source)
	require.False(t, payload.Jobs[2].LogsAvailable)

	require.Equal(t, "job-orphan", payload.Jobs[3].JobID)
	require.Empty(t, payload.Jobs[3].Source)
	require.True(t, payload.Jobs[3].LogsAvailable)
}

func TestPortalLogEntries(t *testing.T) {
	p, dir := testPortal(t)
	writeJobLog(t, dir, "job-1",
		`{"level":"info","ts":1700000000.5,"caller":"executor/run.go:42","msg":"starting","component":"api","attempt":1}`,
		`not json at all`,
		`{"level":"error","ts":1700000001,"msg":"failed"}`,
	)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-1", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		JobID     string     `json:"job_id"`
		Total     int        `json:"total"`
		Truncated bool       `json:"truncated"`
		Entries   []logEntry `json:"entries"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, "job-1", payload.JobID)
	require.Equal(t, 3, payload.Total)
	require.False(t, payload.Truncated)
	require.Len(t, payload.Entries, 3)

	first := payload.Entries[0]
	require.Equal(t, "info", first.Level)
	require.Equal(t, "starting", first.Msg)
	require.NotNil(t, first.Time)
	require.Equal(t, int64(1700000000500), first.Time.UnixMilli())
	require.Equal(t, "api", first.Fields["component"])
	require.NotContains(t, first.Fields, "caller")

	require.Equal(t, "not json at all", payload.Entries[1].Raw)
	require.Empty(t, payload.Entries[1].Msg)

	require.Equal(t, "error", payload.Entries[2].Level)
}

func TestPortalLogTail(t *testing.T) {
	p, dir := testPortal(t)
	lines := make([]string, 0, 30)
	for i := 0; i < 30; i++ {
		lines = append(lines, fmt.Sprintf(`{"level":"info","ts":%d,"msg":"line %d"}`, 1700000000+i, i))
	}
	writeJobLog(t, dir, "job-tail", lines...)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-tail?tail=10", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Total     int        `json:"total"`
		Truncated bool       `json:"truncated"`
		Entries   []logEntry `json:"entries"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, 30, payload.Total)
	require.True(t, payload.Truncated)
	require.Len(t, payload.Entries, 10)
	require.Equal(t, "line 20", payload.Entries[0].Msg)
	require.Equal(t, "line 29", payload.Entries[9].Msg)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-tail?tail=0", nil)
	response = httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestPortalReadsRotatedOTELLogsByJob(t *testing.T) {
	p, dir := testPortal(t)
	writeOTELJobLogs(t, dir, "otel-2026-08-24T10-00-00.000",
		otelJobLogRecord("job-1", "1700000002000000000", "ERROR", "later"),
		otelJobLogRecord("job-other", "1700000001000000000", "INFO", "other job"),
	)
	writeOTELJobLogs(t, dir, "otel",
		otelJobLogRecord("job-1", "1700000000500000000", "INFO", "earlier"),
	)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-1", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Total   int        `json:"total"`
		Entries []logEntry `json:"entries"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	require.Equal(t, 2, payload.Total)
	require.Equal(t, "earlier", payload.Entries[0].Msg)
	require.Equal(t, "info", payload.Entries[0].Level)
	require.Equal(t, "api", payload.Entries[0].Fields["component"])
	require.Equal(t, "runner", payload.Entries[0].Fields["service.name"])
	require.Equal(t, "later", payload.Entries[1].Msg)
	require.NotContains(t, response.Body.String(), "other job")
}

func TestPortalIgnoresPartialActiveOTELBatch(t *testing.T) {
	entries, err := parseOTELLogEntries([]byte(`{"resourceLogs":[`), "job-1")
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestPortalLogRawDownload(t *testing.T) {
	p, dir := testPortal(t)
	writeJobLog(t, dir, "job-raw", `{"level":"info","ts":1700000000,"msg":"hello"}`)

	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-raw?raw=1", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "application/x-ndjson", response.Header().Get("Content-Type"))
	require.Contains(t, response.Header().Get("Content-Disposition"), `job-raw.ndjson`)
	require.Contains(t, response.Body.String(), `"msg":"hello"`)
}

func TestPortalLogValidationAndMissing(t *testing.T) {
	p, _ := testPortal(t)
	h := p.handler()

	for _, id := range []string{"a.b", "%2e%2e%2fsecrets", strings.Repeat("a", 201)} {
		request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/"+id, nil)
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code, id)
	}

	// ServeMux path cleaning intercepts traversal segments before the handler.
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/..", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusMovedPermanently, response.Code)

	request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/logs/job-missing", nil)
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}
