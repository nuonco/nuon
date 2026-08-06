package airgap

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/robfig/cron"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

type Envelope struct {
	Version   string          `json:"version"`
	OrgID     string          `json:"org_id"`
	AppID     string          `json:"app_id"`
	InstallID string          `json:"install_id"`
	CreatedAt time.Time       `json:"created_at"`
	Source    string          `json:"source"`
	AppConfig json.RawMessage `json:"app_config,omitempty"`
	Inputs    []InputSpec     `json:"inputs,omitempty"`
	// ForceDefaultCloudAuth rewrites every cloud-auth block in the composite
	// plans to use the runner process's ambient credentials. Exported plans
	// reference roles provisioned by the sandbox, which don't exist yet when
	// an offline run starts from scratch.
	ForceDefaultCloudAuth bool            `json:"force_default_cloud_auth,omitempty"`
	Components            []ComponentSpec `json:"components,omitempty"`
	Steps                 []Step          `json:"steps"`
	// Day-2 templates are immutable execution templates, not saved plans:
	// every run instantiates a fresh runtime job from the pinned plan.
	Actions  []ActionTemplate  `json:"actions,omitempty"`
	Drift    []DriftTemplate   `json:"drift,omitempty"`
	Runbooks []RunbookTemplate `json:"runbooks,omitempty"`
	// OutputBindings map component-output placeholder tokens baked into
	// composite plans to the producing step's terraform outputs, letting the
	// runner render cross-component references after the producer applies.
	OutputBindings []OutputBinding `json:"output_bindings,omitempty"`
}

// OutputBinding defers one cross-component output reference to execution
// time. Token is the exact placeholder string compiled into downstream
// plans; StepID names the producing component's apply step, whose recorded
// outputs are walked with the dotted OutputPath to yield the value.
type OutputBinding struct {
	Token         string `json:"token"`
	ComponentName string `json:"component_name"`
	StepID        string `json:"step_id"`
	OutputPath    string `json:"output_path"`
}

// ActionTemplate is a self-contained offline action: its composite plan
// carries an ActionWorkflowRunPlan whose steps are fully interpolated inline
// commands (no git sources, no secret params). Publish rejects anything else.
type ActionTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// CronSchedule is a five-field cron expression evaluated in UTC. Empty
	// means manual dispatch only.
	CronSchedule  string          `json:"cron_schedule,omitempty"`
	JobType       string          `json:"job_type"`
	JobGroup      string          `json:"job_group"`
	JobOperation  string          `json:"job_operation"`
	CompositePlan json.RawMessage `json:"composite_plan"`
}

// DriftTemplate is a plan-only terraform run derived from a component's
// deploy plan with any pre-rendered plan contents cleared, so every drift run
// plans fresh against current state. Drift means "drifted from the bundle's
// frozen desired config".
type DriftTemplate struct {
	ID            string          `json:"id"`
	ComponentID   string          `json:"component_id,omitempty"`
	ComponentName string          `json:"component_name"`
	JobType       string          `json:"job_type"`
	JobGroup      string          `json:"job_group"`
	JobOperation  string          `json:"job_operation"`
	CompositePlan json.RawMessage `json:"composite_plan"`
}

const (
	RunbookStepKindAction     = "action"
	RunbookStepKindDrift      = "drift"
	RunbookStepKindHealthGate = "health-gate"
)

// RunbookTemplate is an ordered list of refs to the other templates plus
// health gates; execution stops on the first failure. No branching,
// approvals, retries, event waits, or output interpolation.
type RunbookTemplate struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Steps []RunbookStep `json:"steps"`
}

type RunbookStep struct {
	Kind  string `json:"kind"`
	RefID string `json:"ref_id,omitempty"`
	// Component scopes a health-gate to one component by name; empty gates
	// on every component's health.
	Component string `json:"component,omitempty"`
}

// ComponentSpec is the component-health ownership metadata the connected
// runner would fetch from ctl-api (GetRunnerInstallComponents). It is frozen
// into the envelope at export so the offline health engine can map cluster
// and terraform resources back to components. It carries no credentials.
type ComponentSpec struct {
	InstallComponentID string `json:"install_component_id"`
	ComponentID        string `json:"component_id"`
	ComponentName      string `json:"component_name"`
	ComponentType      string `json:"component_type"`
	HelmReleaseName    string `json:"helm_release_name,omitempty"`
	HelmNamespace      string `json:"helm_namespace,omitempty"`
}

type InputSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
	// Bindable reports whether plan export replaced this input's reference
	// value with a placeholder, so a customer-supplied value takes effect
	// offline. Non-bindable inputs keep the value baked at publish time.
	Bindable bool `json:"bindable,omitempty"`
}

type Step struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	JobType      string   `json:"job_type"`
	JobOperation string   `json:"job_operation"`
	JobGroup     string   `json:"job_group"`
	DependsOn    []string `json:"depends_on,omitempty"`
	// PlanFromStep names an earlier step whose execution result (a rendered
	// plan, e.g. a tfplan) is injected into this step's composite plan as
	// apply_plan_contents at fetch time. This replaces the control plane's
	// plan-approval chaining for offline runs.
	PlanFromStep  string          `json:"plan_from_step,omitempty"`
	CompositePlan json.RawMessage `json:"composite_plan"`
}

func Load(path string) (*Envelope, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read envelope: %w", err)
	}
	return Parse(b)
}

func Parse(b []byte) (*Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(b, &envelope); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	return &envelope, nil
}

func (e *Envelope) Validate() error {
	if e.Version != "v0" {
		return fmt.Errorf("unsupported envelope version %q", e.Version)
	}
	steps := make(map[string]Step, len(e.Steps))
	for _, step := range e.Steps {
		if step.ID == "" {
			return fmt.Errorf("step ID is required")
		}
		if _, ok := steps[step.ID]; ok {
			return fmt.Errorf("duplicate step ID %q", step.ID)
		}
		if err := models.AppRunnerJobType(step.JobType).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("step %q has invalid job type %q: %w", step.ID, step.JobType, err)
		}
		if err := models.AppRunnerJobGroup(step.JobGroup).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("step %q has invalid job group %q: %w", step.ID, step.JobGroup, err)
		}
		steps[step.ID] = step
	}
	for _, step := range e.Steps {
		for _, dependency := range step.DependsOn {
			if _, ok := steps[dependency]; !ok {
				return fmt.Errorf("step %q depends on unknown step %q", step.ID, dependency)
			}
		}
		if step.PlanFromStep != "" {
			if _, ok := steps[step.PlanFromStep]; !ok {
				return fmt.Errorf("step %q takes its plan from unknown step %q", step.ID, step.PlanFromStep)
			}
			if step.PlanFromStep == step.ID {
				return fmt.Errorf("step %q cannot take its plan from itself", step.ID)
			}
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return fmt.Errorf("dependency cycle includes step %q", id)
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true
		for _, dependency := range steps[id].DependsOn {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range steps {
		if err := visit(id); err != nil {
			return err
		}
	}
	tokens := map[string]bool{}
	for _, binding := range e.OutputBindings {
		if binding.Token == "" || binding.ComponentName == "" || binding.StepID == "" || binding.OutputPath == "" {
			return fmt.Errorf("output binding for component %q is incomplete", binding.ComponentName)
		}
		if tokens[binding.Token] {
			return fmt.Errorf("duplicate output binding token %q", binding.Token)
		}
		tokens[binding.Token] = true
		if _, ok := steps[binding.StepID]; !ok {
			return fmt.Errorf("output binding %q references unknown step %q", binding.Token, binding.StepID)
		}
	}
	return e.validateDay2()
}

func (e *Envelope) validateDay2() error {
	refs := map[string]string{}
	for _, action := range e.Actions {
		if action.ID == "" {
			return fmt.Errorf("action template ID is required")
		}
		if _, ok := refs[action.ID]; ok {
			return fmt.Errorf("duplicate day-2 template ID %q", action.ID)
		}
		refs[action.ID] = RunbookStepKindAction
		if err := models.AppRunnerJobType(action.JobType).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("action template %q has invalid job type %q: %w", action.ID, action.JobType, err)
		}
		if err := models.AppRunnerJobGroup(action.JobGroup).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("action template %q has invalid job group %q: %w", action.ID, action.JobGroup, err)
		}
		if len(action.CompositePlan) == 0 {
			return fmt.Errorf("action template %q has no composite plan", action.ID)
		}
		if action.CronSchedule != "" {
			if _, err := cron.ParseStandard(action.CronSchedule); err != nil {
				return fmt.Errorf("action template %q has invalid cron schedule %q: %w", action.ID, action.CronSchedule, err)
			}
		}
	}
	for _, drift := range e.Drift {
		if drift.ID == "" {
			return fmt.Errorf("drift template ID is required")
		}
		if _, ok := refs[drift.ID]; ok {
			return fmt.Errorf("duplicate day-2 template ID %q", drift.ID)
		}
		refs[drift.ID] = RunbookStepKindDrift
		if err := models.AppRunnerJobType(drift.JobType).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("drift template %q has invalid job type %q: %w", drift.ID, drift.JobType, err)
		}
		if err := models.AppRunnerJobGroup(drift.JobGroup).Validate(strfmt.Default); err != nil {
			return fmt.Errorf("drift template %q has invalid job group %q: %w", drift.ID, drift.JobGroup, err)
		}
		if len(drift.CompositePlan) == 0 {
			return fmt.Errorf("drift template %q has no composite plan", drift.ID)
		}
	}
	for _, runbook := range e.Runbooks {
		if runbook.ID == "" {
			return fmt.Errorf("runbook template ID is required")
		}
		if _, ok := refs[runbook.ID]; ok {
			return fmt.Errorf("duplicate day-2 template ID %q", runbook.ID)
		}
		refs[runbook.ID] = "runbook"
		if len(runbook.Steps) == 0 {
			return fmt.Errorf("runbook %q has no steps", runbook.ID)
		}
		for i, step := range runbook.Steps {
			switch step.Kind {
			case RunbookStepKindAction, RunbookStepKindDrift:
				if kind, ok := refs[step.RefID]; !ok || kind != step.Kind {
					return fmt.Errorf("runbook %q step %d references unknown %s template %q", runbook.ID, i, step.Kind, step.RefID)
				}
			case RunbookStepKindHealthGate:
			default:
				return fmt.Errorf("runbook %q step %d has unsupported kind %q", runbook.ID, i, step.Kind)
			}
		}
	}
	return nil
}

func (e *Envelope) FindAction(id string) *ActionTemplate {
	for i := range e.Actions {
		if e.Actions[i].ID == id {
			return &e.Actions[i]
		}
	}
	return nil
}

func (e *Envelope) FindDrift(id string) *DriftTemplate {
	for i := range e.Drift {
		if e.Drift[i].ID == id {
			return &e.Drift[i]
		}
	}
	return nil
}

func (e *Envelope) FindRunbook(id string) *RunbookTemplate {
	for i := range e.Runbooks {
		if e.Runbooks[i].ID == id {
			return &e.Runbooks[i]
		}
	}
	return nil
}
