package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operationstate"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

type portalRun struct {
	RunID           string                     `json:"run_id"`
	DispatchID      string                     `json:"dispatch_id,omitempty"`
	RefID           string                     `json:"ref_id"`
	RefKind         string                     `json:"ref_kind"`
	RefName         string                     `json:"ref_name"`
	Source          string                     `json:"source"`
	Status          string                     `json:"status"`
	Error           string                     `json:"error,omitempty"`
	BundleDigest    string                     `json:"bundle_digest,omitempty"`
	PreviousRunID   string                     `json:"previous_run_id,omitempty"`
	StartedAt       time.Time                  `json:"started_at"`
	FinishedAt      *time.Time                 `json:"finished_at,omitempty"`
	Steps           []portalRunStep            `json:"steps"`
	ResultDirective statestore.ResultDirective `json:"result_directive,omitempty"`
	Events          []statestore.StatusEvent   `json:"events,omitempty"`
}

type portalRunStep struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Kind            string                     `json:"kind"`
	JobID           string                     `json:"job_id,omitempty"`
	Status          string                     `json:"status"`
	Error           string                     `json:"error,omitempty"`
	StartedAt       *time.Time                 `json:"started_at,omitempty"`
	FinishedAt      *time.Time                 `json:"finished_at,omitempty"`
	SourceRunID     string                     `json:"source_run_id,omitempty"`
	ResultDirective statestore.ResultDirective `json:"result_directive,omitempty"`
	Description     string                     `json:"status_description,omitempty"`
	Drift           *operation.DriftResult     `json:"drift,omitempty"`
}

func operationPortalRun(run operation.RunStatus) portalRun {
	result := portalRun{
		RunID: run.RunID, DispatchID: run.DispatchID, RefID: run.RefID, RefKind: run.RefKind,
		RefName: run.RefName, Source: run.Source, Status: run.Status, Error: run.Error,
		BundleDigest: run.BundleDigest, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt, Steps: make([]portalRunStep, 0, len(run.Steps)),
	}
	for _, step := range run.Steps {
		result.Steps = append(result.Steps, portalRunStep{
			ID: step.ID, Name: step.Name, Kind: step.Kind, JobID: step.JobID, Status: step.Status,
			Error: step.Error, StartedAt: step.StartedAt, FinishedAt: step.FinishedAt, Drift: step.Drift,
		})
	}
	return result
}

func bootstrapPortalRun(status statestore.Status, inferredType, inferredPrevious string) portalRun {
	runType := status.RunType
	if runType == "" {
		runType = inferredType
	}
	previous := status.PreviousRunID
	if previous == "" {
		previous = inferredPrevious
	}
	name := "Initial installation"
	if runType == statestore.RunTypeUpgrade {
		name = "Bundle upgrade"
	} else if status.Status == statestore.RunStatusFailed {
		name = "Installation attempt"
	}
	result := portalRun{
		RunID: status.RunID, RefID: runType, RefKind: runType, RefName: name, Source: "bundle",
		Status: status.Status, BundleDigest: status.BundleDigest, PreviousRunID: previous,
		ResultDirective: status.ResultDirective,
		StartedAt:       status.StartedAt, FinishedAt: status.FinishedAt, Steps: make([]portalRunStep, 0, len(status.Steps)),
	}
	if status.FailedStep != "" {
		result.Error = "failed at " + status.FailedStep
	}
	for _, step := range status.Steps {
		stepStatus := step.Status
		description := ""
		if runType == statestore.RunTypeUpgrade && step.StartedAt == nil && step.FinishedAt == nil && stepStatus == "finished" {
			stepStatus = "unknown"
			description = "Prior execution record unavailable"
		}
		jobID := step.ID
		if stepStatus == "auto-skipped" || stepStatus == "unknown" {
			jobID = ""
		}
		result.Steps = append(result.Steps, portalRunStep{
			ID: step.ID, Name: step.Name, Kind: "install-step", JobID: jobID, Status: stepStatus,
			Error: step.Error, StartedAt: step.StartedAt, FinishedAt: step.FinishedAt,
			SourceRunID: step.SourceRunID, Description: description,
			ResultDirective: step.ResultDirective,
		})
	}
	return result
}

func readStatusObject(r *http.Request, store operationstate.State, key string) (*statestore.Status, error) {
	raw, ok, err := store.Get(r.Context(), key)
	if err != nil || !ok {
		return nil, err
	}
	var status statestore.Status
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("decode %s: %w", key, err)
	}
	if status.RunID == "" || status.StartedAt.IsZero() {
		return nil, nil
	}
	return &status, nil
}

func (p *portalServer) bootstrapStatuses(r *http.Request) ([]statestore.Status, error) {
	byID := map[string]statestore.Status{}
	add := func(store operationstate.State, key string) error {
		status, err := readStatusObject(r, store, key)
		if err != nil || status == nil {
			return err
		}
		if _, found := byID[status.RunID]; !found {
			byID[status.RunID] = *status
		}
		return nil
	}
	if err := add(p.store, "status.json"); err != nil {
		return nil, err
	}
	keys, err := p.store.List(r.Context(), statestore.InstallRunsPrefix)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if strings.Contains(key, "/events/") && strings.HasSuffix(key, ".json") {
			raw, ok, readErr := p.store.Get(r.Context(), key)
			if readErr != nil {
				return nil, readErr
			}
			if ok {
				var event statestore.StatusEvent
				if err := json.Unmarshal(raw, &event); err != nil {
					return nil, fmt.Errorf("decode %s: %w", key, err)
				}
				if current, found := byID[event.Status.RunID]; !found || event.CreatedAt.After(current.HeartbeatAt) {
					byID[event.Status.RunID] = event.Status
				}
			}
		} else if strings.HasSuffix(key, "/status.json") {
			if err := add(p.store, key); err != nil {
				return nil, err
			}
		}
	}
	keys, err = p.stackStore.List(r.Context(), "state/")
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if !strings.HasSuffix(key, "/status.json") || strings.Contains(key, "/runs/") {
			continue
		}
		if err := add(p.stackStore, key); err != nil {
			return nil, err
		}
	}
	statuses := make([]statestore.Status, 0, len(byID))
	for _, status := range byID {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].StartedAt.Before(statuses[j].StartedAt) })
	return statuses, nil
}

func (p *portalServer) listRuns(r *http.Request) ([]portalRun, error) {
	byID := map[string]portalRun{}
	keys, err := p.store.List(r.Context(), operation.RunsPrefix)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if !strings.HasSuffix(key, "/status.json") {
			continue
		}
		raw, ok, err := p.store.Get(r.Context(), key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var run operation.RunStatus
		if err := json.Unmarshal(raw, &run); err != nil {
			return nil, fmt.Errorf("decode %s: %w", key, err)
		}
		byID[run.RunID] = operationPortalRun(run)
	}
	statuses, err := p.bootstrapStatuses(r)
	if err != nil {
		return nil, err
	}
	for i, status := range statuses {
		runType, previous := statestore.RunTypeInstall, ""
		if i > 0 {
			runType = statestore.RunTypeUpgrade
			if statuses[i-1].Status == statestore.RunStatusFinished {
				previous = statuses[i-1].RunID
			}
		}
		byID[status.RunID] = bootstrapPortalRun(status, runType, previous)
	}
	if len(statuses) > 0 && p.installStackName != "" && p.installStackReader != nil {
		if stack, err := p.installStackReader.Read(r.Context(), p.installStackName); err == nil {
			initial := byID[statuses[0].RunID]
			stepStatus := stack.Phase
			if stepStatus == "pending" {
				stepStatus = "available"
			}
			initial.Steps = append([]portalRunStep{{
				ID: "install-stack", Name: "Provision install stack", Kind: "cloudformation", Status: stepStatus,
				StartedAt: stack.StartedAt, FinishedAt: stack.UpdatedAt,
			}}, initial.Steps...)
			if stack.StartedAt != nil && stack.StartedAt.Before(initial.StartedAt) {
				initial.StartedAt = *stack.StartedAt
			}
			byID[initial.RunID] = initial
		}
	}
	runs := make([]portalRun, 0, len(byID))
	for _, run := range byID {
		runs = append(runs, run)
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].StartedAt.After(runs[j].StartedAt) })
	return runs, nil
}

func (p *portalServer) readRun(r *http.Request, id string) (*portalRun, error) {
	runs, err := p.listRuns(r)
	if err != nil {
		return nil, err
	}
	for _, run := range runs {
		if run.RunID == id {
			keys, listErr := p.store.List(r.Context(), statestore.InstallRunEventsPrefix(id))
			if listErr != nil {
				return nil, listErr
			}
			for _, key := range keys {
				raw, ok, getErr := p.store.Get(r.Context(), key)
				if getErr != nil {
					return nil, getErr
				}
				if !ok {
					continue
				}
				var event statestore.StatusEvent
				if err := json.Unmarshal(raw, &event); err != nil {
					return nil, err
				}
				run.Events = append(run.Events, event)
			}
			sort.Slice(run.Events, func(i, j int) bool { return run.Events[i].Sequence < run.Events[j].Sequence })
			return &run, nil
		}
	}
	return nil, nil
}
