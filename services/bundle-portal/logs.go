package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
	"go.opentelemetry.io/collector/pdata/plog"
)

const (
	jobLogsSuffix      = ".ndjson"
	otelLogsSuffix     = ".jsonl"
	defaultLogTail     = 1000
	maxLogTail         = 5000
	logSourceInstall   = "install"
	logSourceOperation = "operation"
)

// Job IDs become API path values and legacy state keys, so keep them stricter
// than dispatch IDs and exclude dots.
var jobIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,200}$`)

type jobLogSummary struct {
	JobID         string     `json:"job_id"`
	Source        string     `json:"source,omitempty"`
	Name          string     `json:"name,omitempty"`
	Status        string     `json:"status,omitempty"`
	RunID         string     `json:"run_id,omitempty"`
	RefName       string     `json:"ref_name,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	LogsAvailable bool       `json:"logs_available"`
}

type logEntry struct {
	Time   *time.Time     `json:"time,omitempty"`
	Level  string         `json:"level,omitempty"`
	Msg    string         `json:"msg,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
	Raw    string         `json:"raw,omitempty"`
}

func (p *portalServer) logs(w http.ResponseWriter, r *http.Request) {
	keys, err := p.store.List(r.Context(), statestore.JobLogsPrefix)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	summaries := p.jobSummaries(r)
	available := make(map[string]bool, len(keys))
	otelLogsAvailable := false
	for _, key := range keys {
		if strings.HasSuffix(key, otelLogsSuffix) {
			otelLogsAvailable = true
			continue
		}
		if !strings.HasSuffix(key, jobLogsSuffix) {
			continue
		}
		jobID := strings.TrimSuffix(strings.TrimPrefix(key, statestore.JobLogsPrefix), jobLogsSuffix)
		available[jobID] = true
		if _, ok := summaries[jobID]; !ok {
			summaries[jobID] = jobLogSummary{JobID: jobID}
		}
	}
	jobs := make([]jobLogSummary, 0, len(summaries))
	for jobID, summary := range summaries {
		summary.LogsAvailable = available[jobID] || otelLogsAvailable
		jobs = append(jobs, summary)
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		a, b := jobs[i].StartedAt, jobs[j].StartedAt
		switch {
		case a != nil && b != nil && !a.Equal(*b):
			return a.After(*b)
		case a != nil && b == nil:
			return true
		case a == nil && b != nil:
			return false
		}
		return jobs[i].JobID < jobs[j].JobID
	})
	writeJSON(w, map[string]any{"jobs": jobs})
}

// jobSummaries indexes every known job ID from installation status and
// operation run records; log files without a matching record stay bare.
func (p *portalServer) jobSummaries(r *http.Request) map[string]jobLogSummary {
	summaries := map[string]jobLogSummary{}
	if raw, ok, err := p.store.Get(r.Context(), "status.json"); err == nil && ok {
		var status statestore.Status
		if err := json.Unmarshal(raw, &status); err == nil {
			for _, step := range status.Steps {
				summaries[step.ID] = jobLogSummary{
					JobID:     step.ID,
					Source:    logSourceInstall,
					Name:      step.Name,
					Status:    step.Status,
					StartedAt: step.StartedAt,
				}
			}
		}
	}
	runs, err := p.listRuns(r)
	if err != nil {
		return summaries
	}
	for _, run := range runs {
		for _, step := range run.Steps {
			if step.JobID == "" {
				continue
			}
			startedAt := step.StartedAt
			if startedAt == nil {
				started := run.StartedAt
				startedAt = &started
			}
			summary, found := summaries[step.JobID]
			if !found {
				summary = jobLogSummary{JobID: step.JobID, Source: logSourceOperation}
			}
			summary.Name = step.Name
			summary.Status = step.Status
			summary.RunID = run.RunID
			summary.RefName = run.RefName
			summary.StartedAt = startedAt
			summaries[step.JobID] = summary
		}
	}
	return summaries
}

func (p *portalServer) log(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if !jobIDPattern.MatchString(jobID) {
		writeAPIError(w, fmt.Errorf("job ID %q must be 1-200 characters of [a-zA-Z0-9_-]", jobID), http.StatusBadRequest)
		return
	}
	entries, ok, err := p.readJobLogEntries(r.Context(), jobID)
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !ok {
		writeAPIError(w, fmt.Errorf("no logs found for job %s", jobID), http.StatusNotFound)
		return
	}
	if r.URL.Query().Get("raw") == "1" {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", jobID+jobLogsSuffix))
		for _, entry := range entries {
			raw, err := json.Marshal(entry)
			if err != nil {
				writeAPIError(w, err, http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(append(raw, '\n'))
		}
		return
	}
	tail := defaultLogTail
	if v := r.URL.Query().Get("tail"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 1 || parsed > maxLogTail {
			writeAPIError(w, fmt.Errorf("tail must be an integer between 1 and %d", maxLogTail), http.StatusBadRequest)
			return
		}
		tail = parsed
	}
	total := len(entries)
	if total > tail {
		entries = entries[total-tail:]
	}
	writeJSON(w, map[string]any{
		"job_id":    jobID,
		"total":     total,
		"truncated": total > len(entries),
		"entries":   entries,
	})
}

func (p *portalServer) readJobLogEntries(ctx context.Context, jobID string) ([]logEntry, bool, error) {
	entries := make([]logEntry, 0)
	legacy, foundLegacy, err := p.store.Get(ctx, statestore.JobLogKey(jobID))
	if err != nil {
		return nil, false, err
	}
	if foundLegacy {
		entries = append(entries, parseLogEntries(legacy)...)
	}
	keys, err := p.store.List(ctx, statestore.JobLogsPrefix)
	if err != nil {
		return nil, false, err
	}
	for _, key := range keys {
		if !strings.HasSuffix(key, otelLogsSuffix) {
			continue
		}
		raw, found, err := p.store.Get(ctx, key)
		if err != nil {
			return nil, false, err
		}
		if !found {
			continue
		}
		parsed, err := parseOTELLogEntries(raw, jobID)
		if err != nil {
			return nil, false, fmt.Errorf("decode OTEL log object %s: %w", key, err)
		}
		entries = append(entries, parsed...)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Time == nil || entries[j].Time == nil {
			return false
		}
		return entries[i].Time.Before(*entries[j].Time)
	})
	return entries, len(entries) > 0, nil
}

func parseOTELLogEntries(raw []byte, jobID string) ([]logEntry, error) {
	lines := strings.Split(string(raw), "\n")
	entries := make([]logEntry, 0)
	unmarshaler := &plog.JSONUnmarshaler{}
	for index, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		logs, err := unmarshaler.UnmarshalLogs([]byte(line))
		if err != nil {
			if index == len(lines)-1 && !strings.HasSuffix(string(raw), "\n") {
				continue
			}
			return nil, err
		}
		resourceLogs := logs.ResourceLogs()
		for i := 0; i < resourceLogs.Len(); i++ {
			resource := resourceLogs.At(i)
			resourceFields := resource.Resource().Attributes().AsRaw()
			scopeLogs := resource.ScopeLogs()
			for j := 0; j < scopeLogs.Len(); j++ {
				records := scopeLogs.At(j).LogRecords()
				for k := 0; k < records.Len(); k++ {
					record := records.At(k)
					fields := make(map[string]any, len(resourceFields)+record.Attributes().Len())
					for key, value := range resourceFields {
						fields[key] = value
					}
					for key, value := range record.Attributes().AsRaw() {
						fields[key] = value
					}
					recordJobID, _ := fields["runner_job.id"].(string)
					if recordJobID != jobID {
						continue
					}
					delete(fields, "runner_job.id")
					timestamp := record.Timestamp()
					if timestamp == 0 {
						timestamp = record.ObservedTimestamp()
					}
					var observedAt *time.Time
					if timestamp != 0 {
						value := timestamp.AsTime().UTC()
						observedAt = &value
					}
					message, ok := record.Body().AsRaw().(string)
					if !ok {
						message = fmt.Sprint(record.Body().AsRaw())
					}
					if len(fields) == 0 {
						fields = nil
					}
					entries = append(entries, logEntry{Time: observedAt, Level: strings.ToLower(record.SeverityText()), Msg: message, Fields: fields})
				}
			}
		}
	}
	return entries, nil
}

func parseLogEntries(raw []byte) []logEntry {
	lines := strings.Split(string(raw), "\n")
	entries := make([]logEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entries = append(entries, parseLogLine(line))
	}
	return entries
}

func parseLogLine(line string) logEntry {
	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return logEntry{Raw: line}
	}
	entry := logEntry{}
	if ts, ok := record["ts"].(float64); ok {
		when := time.UnixMilli(int64(ts * 1000)).UTC()
		entry.Time = &when
	}
	entry.Level, _ = record["level"].(string)
	entry.Msg, _ = record["msg"].(string)
	for _, key := range []string{"ts", "level", "msg", "caller"} {
		delete(record, key)
	}
	if len(record) > 0 {
		entry.Fields = record
	}
	return entry
}
