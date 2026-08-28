package operationrun

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

type Executor struct {
	client   *customermanaged.Client
	envelope *customermanaged.Envelope
	store    statestore.Store
	loader   CandidateBundleLoader
	source   *customermanaged.BundleSource
	mu       sync.Mutex
	busy     bool
}

type CandidateBundle struct {
	Envelope *customermanaged.Envelope
	Source   *customermanaged.BundleSource
	Close    func() error
}

type CandidateBundleLoader interface {
	Load(context.Context, operation.Request) (*CandidateBundle, error)
}

func NewExecutor(client *customermanaged.Client, envelope *customermanaged.Envelope, store statestore.Store) *Executor {
	return &Executor{client: client, envelope: envelope, store: store}
}

func NewExecutorWithCandidateLoader(client *customermanaged.Client, envelope *customermanaged.Envelope, store statestore.Store, source *customermanaged.BundleSource, loader CandidateBundleLoader) *Executor {
	return &Executor{client: client, envelope: envelope, store: store, source: source, loader: loader}
}

func (e *Executor) Busy() bool { e.mu.Lock(); defer e.mu.Unlock(); return e.busy }

func (e *Executor) Execute(ctx context.Context, request operation.Request, runID string) (*operation.RunStatus, error) {
	e.mu.Lock()
	if e.busy {
		e.mu.Unlock()
		return nil, fmt.Errorf("operation executor is busy")
	}
	e.busy = true
	e.mu.Unlock()
	defer func() { e.mu.Lock(); e.busy = false; e.mu.Unlock() }()
	if request.RefKind == operation.RefKindBundlePlan {
		return e.executeBundlePlan(ctx, request, runID)
	}

	run := &operation.RunStatus{RunID: runID, DispatchID: request.DispatchID, RefID: request.RefID, Source: request.Source, Status: operation.RunStatusInProgress, StartedAt: time.Now().UTC()}
	if action := e.envelope.FindAction(request.RefID); action != nil {
		run.RefKind, run.RefName = operation.RefKindAction, action.Name
		run.Steps = []operation.RunStep{{ID: action.ID, Name: action.Name, Kind: operation.RefKindAction, Status: operation.RunStatusInProgress}}
	} else if drift := e.envelope.FindDrift(request.RefID); drift != nil {
		run.RefKind, run.RefName = operation.RefKindDrift, drift.ComponentName
		run.Steps = []operation.RunStep{{ID: drift.ID, Name: drift.ComponentName, Kind: operation.RefKindDrift, Status: operation.RunStatusInProgress}}
	} else if book := e.envelope.FindRunbook(request.RefID); book != nil {
		run.RefKind, run.RefName = operation.RefKindRunbook, book.Name
		for i, step := range book.Steps {
			run.Steps = append(run.Steps, operation.RunStep{ID: fmt.Sprintf("%s-%d", book.ID, i+1), Name: e.stepName(step), Kind: step.Kind, Status: "available"})
		}
	} else {
		return nil, fmt.Errorf("unknown operation ref %q", request.RefID)
	}
	if err := e.persist(run); err != nil {
		return nil, err
	}

	failed := false
	for i := range run.Steps {
		if failed {
			run.Steps[i].Status = operation.StepStatusDiscarded
			continue
		}
		step := &run.Steps[i]
		now := time.Now().UTC()
		step.StartedAt, step.Status = &now, operation.RunStatusInProgress
		err := e.executeStep(ctx, run.RunID, step, request.RefID)
		finished := time.Now().UTC()
		step.FinishedAt = &finished
		if err != nil {
			step.Status, step.Error, failed = operation.RunStatusFailed, err.Error(), true
		} else {
			step.Status = operation.RunStatusFinished
		}
		if err := e.persist(run); err != nil {
			return nil, err
		}
	}
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if failed {
		run.Status = operation.RunStatusFailed
		for _, step := range run.Steps {
			if step.Error != "" {
				run.Error = step.Error
				break
			}
		}
	} else {
		run.Status = operation.RunStatusFinished
	}
	return run, e.persist(run)
}

func (e *Executor) executeBundlePlan(ctx context.Context, request operation.Request, runID string) (*operation.RunStatus, error) {
	if e.loader == nil {
		return nil, fmt.Errorf("candidate bundle loading is not configured")
	}
	candidate, err := e.loader.Load(ctx, request)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, fmt.Errorf("candidate bundle loader returned no bundle")
	}
	if candidate.Close != nil {
		defer candidate.Close()
	}
	if candidate.Envelope == nil || candidate.Envelope.InstallID != request.DeploymentID {
		return nil, fmt.Errorf("candidate deployment ID mismatch")
	}
	if e.source != nil && candidate.Source != nil {
		defer e.source.Overlay(candidate.Source)()
	}

	now := time.Now().UTC()
	run := &operation.RunStatus{
		RunID: runID, DispatchID: request.DispatchID, RefID: request.RefID, RefKind: operation.RefKindBundlePlan,
		RefName: "candidate bundle plan", Source: request.Source, Status: operation.RunStatusInProgress,
		BundleDigest: request.BundleDigest, StartedAt: now,
		Steps: []operation.RunStep{{ID: "install-stack-plan", Name: "install stack plan", Kind: operation.RefKindBundlePlan, Status: operation.RunStatusFinished, StartedAt: &now, FinishedAt: &now}},
	}
	steps := make(map[string]customermanaged.Step, len(candidate.Envelope.Steps))
	for _, step := range candidate.Envelope.Steps {
		steps[step.ID] = step
	}
	for _, id := range request.PlanStepIDs {
		step, found := steps[id]
		if !found {
			return nil, fmt.Errorf("candidate plan step %q not found", id)
		}
		if step.JobOperation != "create-apply-plan" {
			return nil, fmt.Errorf("candidate step %q operation %q is not create-apply-plan", id, step.JobOperation)
		}
		run.Steps = append(run.Steps, operation.RunStep{ID: step.ID, Name: step.Name, Kind: step.JobGroup, Status: "available"})
	}
	if err := e.persist(run); err != nil {
		return nil, err
	}
	failed := false
	for i := 1; i < len(run.Steps); i++ {
		stepStatus := &run.Steps[i]
		step := steps[stepStatus.ID]
		started := time.Now().UTC()
		stepStatus.StartedAt, stepStatus.Status = &started, operation.RunStatusInProgress
		stepStatus.JobID = runID + "--" + step.ID
		handle, jobErr := e.client.EnqueueOperationJobWithEnvelope(stepStatus.JobID, step.JobType, step.JobGroup, step.JobOperation, step.CompositePlan, candidate.Envelope)
		if jobErr == nil {
			jobErr = handle.Await(ctx)
		}
		finished := time.Now().UTC()
		stepStatus.FinishedAt = &finished
		if jobErr != nil {
			stepStatus.Status, stepStatus.Error, failed = operation.RunStatusFailed, jobErr.Error(), true
		} else {
			stepStatus.Status = operation.RunStatusFinished
		}
		if err := e.persist(run); err != nil {
			return nil, err
		}
		if failed {
			break
		}
	}
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if failed {
		run.Status = operation.RunStatusFailed
		for _, step := range run.Steps {
			if step.Error != "" {
				run.Error = step.Error
				break
			}
		}
	} else {
		run.Status = operation.RunStatusFinished
	}
	return run, e.persist(run)
}

func (e *Executor) executeStep(ctx context.Context, runID string, step *operation.RunStep, topRef string) error {
	refID := step.ID
	if e.envelope.FindRunbook(topRef) != nil {
		book := e.envelope.FindRunbook(topRef)
		for i := range book.Steps {
			if step.ID == fmt.Sprintf("%s-%d", book.ID, i+1) {
				refID = book.Steps[i].RefID
				if step.Kind == customermanaged.RunbookStepKindHealthGate {
					return e.healthGate(book.Steps[i].Component)
				}
				break
			}
		}
	}
	if action := e.envelope.FindAction(refID); action != nil {
		step.JobID = runID + "--" + action.ID
		h, err := e.client.EnqueueOperationJob(step.JobID, action.JobType, action.JobGroup, action.JobOperation, action.CompositePlan)
		if err != nil {
			return err
		}
		return h.Await(ctx)
	}
	if drift := e.envelope.FindDrift(refID); drift != nil {
		step.JobID = runID + "--" + drift.ID
		h, err := e.client.EnqueueOperationJob(step.JobID, drift.JobType, drift.JobGroup, drift.JobOperation, drift.CompositePlan)
		if err != nil {
			return err
		}
		if err := h.Await(ctx); err != nil {
			return err
		}
		plan, err := h.PlanJSON()
		if err != nil {
			return err
		}
		// Best-effort: the raw plan is a debugging artifact; classification
		// below is the authoritative verdict and must not fail with it.
		_ = e.store.WriteFile(operation.JobPlanKey(step.JobID), plan)
		result, err := ClassifyDrift(plan)
		step.Drift = result
		return err
	}
	return fmt.Errorf("unknown step ref %q", refID)
}

func (e *Executor) healthGate(component string) error {
	snapshot := e.client.LatestHealth()
	if snapshot == nil {
		return fmt.Errorf("health is unknown")
	}
	found := false
	for _, health := range snapshot.Components {
		if component == "" || health.ComponentName == component {
			found = true
			if health.Health != "healthy" {
				return fmt.Errorf("component %s is %s", health.ComponentName, health.Health)
			}
		}
	}
	if !found {
		return fmt.Errorf("component %q health is unknown", component)
	}
	return nil
}

func (e *Executor) stepName(step customermanaged.RunbookStep) string {
	if action := e.envelope.FindAction(step.RefID); action != nil {
		return action.Name
	}
	if drift := e.envelope.FindDrift(step.RefID); drift != nil {
		return drift.ComponentName
	}
	if step.Component != "" {
		return "health: " + step.Component
	}
	return "health gate"
}

func (e *Executor) persist(run *operation.RunStatus) error {
	b, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	return e.store.WriteFile(operation.RunStatusKey(run.RunID), append(b, '\n'))
}

func ClassifyDrift(planJSON []byte) (*operation.DriftResult, error) {
	var plan tfjson.Plan
	if err := json.Unmarshal(planJSON, &plan); err != nil {
		return nil, err
	}
	result := &operation.DriftResult{}
	actionableResources := make(map[string]bool, len(plan.ResourceChanges))
	for _, change := range plan.ResourceChanges {
		if change != nil && isTerraformChange(change.Change) {
			result.ResourceChanges++
			actionableResources[change.Address] = true
		}
	}
	for _, change := range plan.OutputChanges {
		if isTerraformChange(change) {
			result.OutputChanges++
		}
	}
	driftedResources := make(map[string]bool, len(plan.ResourceDrift))
	for _, drift := range plan.ResourceDrift {
		if drift != nil && isTerraformChange(drift.Change) && actionableResources[drift.Address] {
			result.ResourceDrift++
			driftedResources[drift.Address] = true
		}
	}
	var changed, noops []operation.DriftResourceChange
	for _, change := range plan.ResourceChanges {
		if change == nil {
			continue
		}
		rc := operation.DriftResourceChange{
			Address: change.Address,
			Action:  changeAction(change.Change),
			Drifted: driftedResources[change.Address],
		}
		if rc.Action == "noop" && !rc.Drifted {
			noops = append(noops, rc)
		} else {
			changed = append(changed, rc)
		}
	}
	resources := append(changed, noops...)
	if len(resources) > operation.MaxDriftResources {
		resources = resources[:operation.MaxDriftResources]
		result.ResourcesTruncated = true
	}
	result.Resources = resources
	result.Drifted = result.ResourceChanges > 0 || result.OutputChanges > 0 || result.ResourceDrift > 0 || len(plan.DeferredChanges) > 0
	if result.Drifted {
		result.Summary = fmt.Sprintf("%d resource changes, %d output changes, %d drifted resources", result.ResourceChanges, result.OutputChanges, result.ResourceDrift)
	} else {
		result.Summary = "no drift"
	}
	return result, nil
}

func changeAction(change *tfjson.Change) string {
	if change == nil {
		return "noop"
	}
	switch {
	case change.Actions.Replace():
		return "replace"
	case change.Actions.Create():
		return "create"
	case change.Actions.Delete():
		return "destroy"
	case change.Actions.Update():
		return "update"
	default:
		return "noop"
	}
}

func isTerraformChange(change *tfjson.Change) bool {
	if change == nil {
		return false
	}
	for _, action := range change.Actions {
		if action != tfjson.ActionNoop {
			return true
		}
	}
	return false
}
