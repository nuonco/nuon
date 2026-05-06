package interests

import (
	"encoding/json"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestMatches(t *testing.T) {
	provisionStart := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "provision",
		Phase:        signal.SignalPhaseExecute,
	}
	provisionSuccess := provisionStart
	deprovisionFailed := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "deprovision",
		Phase:        signal.SignalPhaseExecute,
	}
	manualDeploySuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "manual_deploy",
		Phase:        signal.SignalPhaseExecute,
	}
	driftRunSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "drift_run",
		Phase:        signal.SignalPhaseExecute,
	}
	inputUpdateSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "input_update",
		Phase:        signal.SignalPhaseExecute,
	}
	syncSecretsSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "sync_secrets",
		Phase:        signal.SignalPhaseExecute,
	}
	sandboxDriftSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "drift_run_reprovision_sandbox",
		Phase:        signal.SignalPhaseExecute,
	}
	sandboxReprovisionSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "reprovision_sandbox",
		Phase:        signal.SignalPhaseExecute,
	}
	unknownType := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflow,
		WorkflowType: "app_branches_manual_update", // not in the v1 taxonomy
		Phase:        signal.SignalPhaseExecute,
	}
	approvalRequest := signal.SignalPhaseEvent{
		SignalType:   signalTypeWorkflowStepApprovalRequest,
		WorkflowType: "deploy_components",
		Phase:        signal.SignalPhaseExecute,
		StepID:       "iws_step_1",
	}
	// Step events inside drift workflows. Without DB enrichment, these
	// classify via stepResolutionFromParent; with DB enrichment the
	// classification may differ — but suppression keys off the parent
	// WorkflowType so both paths are covered.
	driftRunStepSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflowStep,
		WorkflowType: "drift_run",
		Phase:        signal.SignalPhaseExecute,
	}
	sandboxDriftStepSuccess := signal.SignalPhaseEvent{
		SignalType:   signalTypeExecuteWorkflowStep,
		WorkflowType: "drift_run_reprovision_sandbox",
		Phase:        signal.SignalPhaseExecute,
	}
	approvalResponse := signal.SignalPhaseEvent{
		SignalType:   signalTypeWorkflowStepApprovalResponse,
		WorkflowType: "deploy_components",
		Phase:        signal.SignalPhaseExecute,
		StepID:       "iws_step_1",
	}

	successOutcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}
	failureOutcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusError}
	cancelledOutcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusCancelled}

	cases := []struct {
		name    string
		event   signal.SignalPhaseEvent
		outcome *signal.SignalPhaseOutcome
		in      Interests
		want    bool
	}{
		{
			name:  "AllEvents matches everything",
			event: provisionStart,
			in:    AllEvents(),
			want:  true,
		},
		{
			name:    "AllEvents matches terminal too",
			event:   deprovisionFailed,
			outcome: failureOutcome,
			in:      AllEvents(),
			want:    true,
		},
		{
			name:  "Empty config matches nothing",
			event: provisionStart,
			in:    Interests{},
			want:  false,
		},
		{
			name:  "Resource missing — no match",
			event: provisionStart,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:  "Installs.* matches with empty Ops (all sub-ops)",
			event: provisionStart,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeAll},
			}},
			want: true,
		},
		{
			name:  "Installs.provision matches when only provision is listed",
			event: provisionStart,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Ops: []string{"provision"}, Outcome: OutcomeAll},
			}},
			want: true,
		},
		{
			name:  "Installs.deprovision filtered out by Ops list",
			event: deprovisionFailed,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Ops: []string{"provision"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:  "Outcome=completion drops started events",
			event: provisionStart,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeCompletion},
			}},
			want: false,
		},
		{
			name:    "Outcome=completion keeps succeeded",
			event:   provisionSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeCompletion},
			}},
			want: true,
		},
		{
			name:    "Outcome=completion keeps failed",
			event:   deprovisionFailed,
			outcome: failureOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeCompletion},
			}},
			want: true,
		},
		{
			name:    "Outcome=failures drops succeeded",
			event:   provisionSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeFailures},
			}},
			want: false,
		},
		{
			name:    "Outcome=failures keeps failed",
			event:   deprovisionFailed,
			outcome: failureOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeFailures},
			}},
			want: true,
		},
		{
			name:    "Outcome=failures keeps cancelled",
			event:   deprovisionFailed,
			outcome: cancelledOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeFailures},
			}},
			want: true,
		},
		{
			name:    "Empty Outcome treated as 'all'",
			event:   provisionStart,
			outcome: nil,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: ""},
			}},
			want: true,
		},
		{
			name:    "manual_deploy envelope no longer classifies — does not match installs",
			event:   manualDeploySuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "manual_deploy envelope no longer classifies — does not match components",
			event:   manualDeploySuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			// Drift workflow lifecycle events are unconditionally suppressed
			// — listing "drift" in Ops is a no-op, drift surfaces only via
			// the drift-detected event class gated by DriftDetected.
			name:    "drift_run lifecycle suppressed even when components.Ops contains drift",
			event:   driftRunSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"drift"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "drift_run lifecycle suppressed under AllEvents",
			event:   driftRunSuccess,
			outcome: successOutcome,
			in:      Interests{AllEvents: true},
			want:    false,
		},
		{
			name:    "drift_run envelope does not match components.deploy",
			event:   driftRunSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"deploy"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "input_update matches install_configurations.inputs",
			event:   inputUpdateSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstallConfigurations: {Ops: []string{"inputs"}, Outcome: OutcomeAll},
			}},
			want: true,
		},
		{
			name:    "input_update no longer matches installs",
			event:   inputUpdateSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "sync_secrets matches install_configurations.secrets",
			event:   syncSecretsSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstallConfigurations: {Ops: []string{"secrets"}, Outcome: OutcomeAll},
			}},
			want: true,
		},
		{
			name:    "sandbox drift workflow lifecycle suppressed even when sandboxes.Ops contains drift",
			event:   sandboxDriftSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Ops: []string{"drift"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "sandbox drift workflow lifecycle suppressed under AllEvents",
			event:   sandboxDriftSuccess,
			outcome: successOutcome,
			in:      Interests{AllEvents: true},
			want:    false,
		},
		{
			// Regression: steps inside a drift workflow used to leak through
			// because they classify as their own resource (e.g. a
			// "runner healthy" step → (runners, reprovision)). Suppression
			// keys off the parent WorkflowType so the entire lifecycle tree
			// is dropped — only the dedicated drift-detected event reaches
			// subscribers.
			name:    "drift_run step lifecycle suppressed under AllEvents",
			event:   driftRunStepSuccess,
			outcome: successOutcome,
			in:      Interests{AllEvents: true},
			want:    false,
		},
		{
			name:    "sandbox drift step lifecycle suppressed under AllEvents",
			event:   sandboxDriftStepSuccess,
			outcome: successOutcome,
			in:      Interests{AllEvents: true},
			want:    false,
		},
		{
			name:    "sandbox drift step lifecycle suppressed even when runners.reprovision is requested",
			event:   sandboxDriftStepSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceRunners: {Ops: []string{"reprovision"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "sandbox drift workflow does not match sandboxes.reprovision",
			event:   sandboxDriftSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Ops: []string{"reprovision"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:    "sandbox manual reprovision matches sandboxes.reprovision",
			event:   sandboxReprovisionSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Ops: []string{"reprovision"}, Outcome: OutcomeAll},
			}},
			want: true,
		},
		{
			name:    "sandbox manual reprovision does not match sandboxes.drift",
			event:   sandboxReprovisionSuccess,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Ops: []string{"drift"}, Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:  "Unknown WorkflowType does not match",
			event: unknownType,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceInstalls: {Outcome: OutcomeAll},
			}},
			want: false,
		},
		{
			name:  "Approval request matches when ApprovalRequests=true",
			event: approvalRequest,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {ApprovalRequests: true},
			}},
			want: true,
		},
		{
			name:  "Approval request dropped when ApprovalRequests=false",
			event: approvalRequest,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {ApprovalRequests: false, ApprovalResponses: true},
			}},
			want: false,
		},
		{
			name:    "Approval response matches when ApprovalResponses=true",
			event:   approvalResponse,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {ApprovalResponses: true},
			}},
			want: true,
		},
		{
			name:    "Approval response dropped when ApprovalResponses=false",
			event:   approvalResponse,
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {ApprovalResponses: false, ApprovalRequests: true},
			}},
			want: false,
		},
		{
			name:  "Approval ignores Outcome filter — only ApprovalRequests gates",
			event: approvalRequest,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Outcome: OutcomeFailures, ApprovalRequests: true},
			}},
			want: true,
		},
		{
			name:    "Approval still requires resource match",
			event:   approvalRequest,
			outcome: nil,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {ApprovalRequests: true},
			}},
			want: false,
		},
		{
			name:    "Approval honours Ops filter on the resource",
			event:   approvalRequest,
			outcome: nil,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"teardown"}, ApprovalRequests: true},
			}},
			want: false,
		},
		{
			// Canonical post-suppression shape: just DriftDetected=true with
			// no "drift" in Ops. Drift no longer participates in the SubOps
			// vocabulary — it surfaces exclusively through this event class.
			name: "drift-detected (drift_run) matches components when DriftDetected=true (no drift Op)",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {DriftDetected: true},
			}},
			want: true,
		},
		{
			// Legacy backward-compat: stored configs that still list "drift"
			// in Ops are accepted by the matcher (validate rejects new ones).
			name: "drift-detected (drift_run) matches components.drift legacy Ops shape",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"drift"}, DriftDetected: true},
			}},
			want: true,
		},
		{
			name: "drift-detected (drift_run) dropped when DriftDetected=false",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"drift"}, DriftDetected: false},
			}},
			want: false,
		},
		{
			name: "drift-detected (drift_run) honours Ops filter — does not match deploy",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Ops: []string{"deploy"}, DriftDetected: true},
			}},
			want: false,
		},
		{
			name: "drift-detected (drift_run_reprovision_sandbox) matches sandboxes.drift",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run_reprovision_sandbox",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift_sb",
			},
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceSandboxes: {Ops: []string{"drift"}, DriftDetected: true},
			}},
			want: true,
		},
		{
			name: "drift-detected ignores Outcome filter — matches even when Outcome=failures",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			outcome: successOutcome,
			in: Interests{Resources: map[ResourceKind]ResourceCfg{
				ResourceComponents: {Outcome: OutcomeFailures, DriftDetected: true},
			}},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Matches(tc.event, tc.outcome, nil, tc.in)
			if got != tc.want {
				t.Fatalf("Matches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStepResolution(t *testing.T) {
	cases := []struct {
		name           string
		stepTargetType string
		parentWfType   string
		wantResource   ResourceKind
		wantOp         string
		wantOK         bool
	}{
		{
			name:           "sandbox run inside provision → sandboxes.provision",
			stepTargetType: stepTargetInstallSandboxRun,
			parentWfType:   "provision",
			wantResource:   ResourceSandboxes,
			wantOp:         "provision",
			wantOK:         true,
		},
		{
			name:           "sandbox run inside reprovision_sandbox → sandboxes.reprovision",
			stepTargetType: stepTargetInstallSandboxRun,
			parentWfType:   "reprovision_sandbox",
			wantResource:   ResourceSandboxes,
			wantOp:         "reprovision",
			wantOK:         true,
		},
		{
			name:           "sandbox run inside deprovision_sandbox → sandboxes.deprovision",
			stepTargetType: stepTargetInstallSandboxRun,
			parentWfType:   "deprovision_sandbox",
			wantResource:   ResourceSandboxes,
			wantOp:         "deprovision",
			wantOK:         true,
		},
		{
			name:           "deploy step inside deploy_components → components.deploy",
			stepTargetType: stepTargetInstallDeploy,
			parentWfType:   "deploy_components",
			wantResource:   ResourceComponents,
			wantOp:         "deploy",
			wantOK:         true,
		},
		{
			name:           "deploy step inside teardown_components → components.teardown",
			stepTargetType: stepTargetInstallDeploy,
			parentWfType:   "teardown_components",
			wantResource:   ResourceComponents,
			wantOp:         "teardown",
			wantOK:         true,
		},
		{
			name:           "action workflow run step → actions.run",
			stepTargetType: stepTargetInstallActionWorkflowRun,
			parentWfType:   "provision",
			wantResource:   ResourceActions,
			wantOp:         "run",
			wantOK:         true,
		},
		{
			name:           "runner update step inside provision → runners.provision",
			stepTargetType: stepTargetInstallRunnerUpdate,
			parentWfType:   "provision",
			wantResource:   ResourceRunners,
			wantOp:         "provision",
			wantOK:         true,
		},
		{
			name:           "runner update step inside reprovision_sandbox → runners.reprovision",
			stepTargetType: stepTargetInstallRunnerUpdate,
			parentWfType:   "reprovision_sandbox",
			wantResource:   ResourceRunners,
			wantOp:         "reprovision",
			wantOK:         true,
		},
		{
			name:           "unknown step target type → not resolved",
			stepTargetType: "install_random_thing",
			parentWfType:   "provision",
			wantOK:         false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, op, ok := stepResolution(tc.stepTargetType, tc.parentWfType)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if res != tc.wantResource || op != tc.wantOp {
				t.Fatalf("got (%s, %s), want (%s, %s)", res, op, tc.wantResource, tc.wantOp)
			}
		})
	}
}

func TestClassifySlugs(t *testing.T) {
	cases := []struct {
		name    string
		event   signal.SignalPhaseEvent
		outcome *signal.SignalPhaseOutcome
		want    []string
	}{
		{
			name: "execute-workflow provision started",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeExecuteWorkflow,
				WorkflowType: "provision",
				Phase:        signal.SignalPhaseExecute,
			},
			want: []string{
				"resource:installs",
				"op:installs.provision",
				"event:lifecycle.started",
			},
		},
		{
			name: "execute-workflow provision succeeded",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeExecuteWorkflow,
				WorkflowType: "provision",
				Phase:        signal.SignalPhaseExecute,
			},
			outcome: &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			want: []string{
				"resource:installs",
				"op:installs.provision",
				"event:lifecycle.succeeded",
				"outcome:completion",
			},
		},
		{
			name: "execute-workflow provision failed",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeExecuteWorkflow,
				WorkflowType: "provision",
				Phase:        signal.SignalPhaseExecute,
			},
			outcome: &signal.SignalPhaseOutcome{Status: signal.SignalStatusError},
			want: []string{
				"resource:installs",
				"op:installs.provision",
				"event:lifecycle.failed",
				"outcome:completion",
				"outcome:failures",
			},
		},
		{
			name: "execute-workflow provision cancelled",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeExecuteWorkflow,
				WorkflowType: "provision",
				Phase:        signal.SignalPhaseCancel,
			},
			outcome: &signal.SignalPhaseOutcome{Status: signal.SignalStatusCancelled},
			want: []string{
				"resource:installs",
				"op:installs.provision",
				"event:lifecycle.cancelled",
				"outcome:completion",
				"outcome:failures",
			},
		},
		{
			name: "approval request slugs",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeWorkflowStepApprovalRequest,
				WorkflowType: "deploy_components",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_x",
			},
			want: []string{
				"resource:components",
				"op:components.deploy",
				"event:approval.request",
			},
		},
		{
			name: "approval response slugs (no DB → generic only)",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeWorkflowStepApprovalResponse,
				WorkflowType: "deploy_components",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_x",
			},
			outcome: &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
			want: []string{
				"resource:components",
				"op:components.deploy",
				"event:approval.response",
			},
		},
		{
			name: "unknown workflow type → no slugs",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeExecuteWorkflow,
				WorkflowType: "something_we_dont_know",
				Phase:        signal.SignalPhaseExecute,
			},
			want: nil,
		},
		{
			name: "drift-detected (drift_run) → components.drift slugs",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift",
			},
			want: []string{
				"resource:components",
				"op:components.drift",
				"event:drift.detected",
			},
		},
		{
			name: "drift-detected (drift_run_reprovision_sandbox) → sandboxes.drift slugs",
			event: signal.SignalPhaseEvent{
				SignalType:   signalTypeDriftDetected,
				WorkflowType: "drift_run_reprovision_sandbox",
				Phase:        signal.SignalPhaseExecute,
				StepID:       "iws_drift_sb",
			},
			want: []string{
				"resource:sandboxes",
				"op:sandboxes.drift",
				"event:drift.detected",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Classify(tc.event, tc.outcome, nil)
			if !equalStrings(got, tc.want) {
				t.Fatalf("Classify() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestSlugHelpers(t *testing.T) {
	if got := ResourceSlug(ResourceInstalls); got != "resource:installs" {
		t.Fatalf("ResourceSlug = %q", got)
	}
	if got := OpSlug(ResourceComponents, "deploy"); got != "op:components.deploy" {
		t.Fatalf("OpSlug = %q", got)
	}
}

func TestDefaults(t *testing.T) {
	if !AllEvents().AllEvents {
		t.Fatal("AllEvents() should set AllEvents=true")
	}
	if len(AllEvents().Resources) != 0 {
		t.Fatal("AllEvents() should leave Resources empty")
	}

	d := Default()
	if d.AllEvents {
		t.Fatal("Default() must have AllEvents=false")
	}
	want := []ResourceKind{
		ResourceInstalls, ResourceComponents, ResourceSandboxes, ResourceInstallConfigurations,
	}
	for _, k := range want {
		cfg, ok := d.Resources[k]
		if !ok {
			t.Fatalf("Default() missing resource %q", k)
		}
		if cfg.Outcome != OutcomeCompletion {
			t.Fatalf("Default()[%s].Outcome = %q, want %q", k, cfg.Outcome, OutcomeCompletion)
		}
		if !cfg.ApprovalRequests || !cfg.ApprovalResponses {
			t.Fatalf("Default()[%s] approval flags must both be true", k)
		}
		if len(cfg.Ops) != 0 {
			t.Fatalf("Default()[%s].Ops should be empty (all sub-ops)", k)
		}
	}
	for _, k := range []ResourceKind{ResourceComponents, ResourceSandboxes} {
		if !d.Resources[k].DriftDetected {
			t.Fatalf("Default()[%s].DriftDetected must be true", k)
		}
	}
	for _, k := range []ResourceKind{ResourceInstalls, ResourceInstallConfigurations} {
		if d.Resources[k].DriftDetected {
			t.Fatalf("Default()[%s].DriftDetected must be false (no drift sub-op)", k)
		}
	}
	for _, k := range []ResourceKind{ResourceRunners, ResourceActions} {
		if _, ok := d.Resources[k]; ok {
			t.Fatalf("Default() should not include %q", k)
		}
	}
}

func TestInterestsRoundTripJSON(t *testing.T) {
	in := Interests{
		Resources: map[ResourceKind]ResourceCfg{
			ResourceInstalls: {
				Ops:               []string{"provision", "reprovision"},
				Outcome:           OutcomeFailures,
				ApprovalRequests:  true,
				ApprovalResponses: false,
			},
		},
	}
	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	raw, ok := v.([]byte)
	if !ok {
		t.Fatalf("Value type = %T, want []byte", v)
	}

	// Sanity check: JSON should NOT include all_events when false.
	var asMap map[string]any
	if err := json.Unmarshal(raw, &asMap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := asMap["all_events"]; ok {
		t.Fatalf("all_events should be omitted when false; got %s", string(raw))
	}

	var out Interests
	if err := out.Scan(raw); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if out.AllEvents {
		t.Fatal("AllEvents should be false after round trip")
	}
	cfg, ok := out.Resources[ResourceInstalls]
	if !ok {
		t.Fatal("missing installs after round trip")
	}
	if cfg.Outcome != OutcomeFailures {
		t.Fatalf("Outcome = %q, want %q", cfg.Outcome, OutcomeFailures)
	}
	if !cfg.ApprovalRequests || cfg.ApprovalResponses {
		t.Fatalf("approval flags lost in round trip: %+v", cfg)
	}
	if len(cfg.Ops) != 2 || cfg.Ops[0] != "provision" || cfg.Ops[1] != "reprovision" {
		t.Fatalf("Ops round trip wrong: %+v", cfg.Ops)
	}
}

func TestInterestsScanNullAndEmpty(t *testing.T) {
	var i Interests
	if err := i.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if !i.IsZero() {
		t.Fatal("Scan(nil) should leave Interests zero-valued")
	}

	if err := i.Scan([]byte{}); err != nil {
		t.Fatalf("Scan(empty): %v", err)
	}
	if !i.IsZero() {
		t.Fatal("Scan(empty bytes) should leave Interests zero-valued")
	}

	if err := i.Scan(""); err != nil {
		t.Fatalf("Scan(empty string): %v", err)
	}
	if !i.IsZero() {
		t.Fatal("Scan(empty string) should leave Interests zero-valued")
	}
}

func TestInterestsValueZeroIsNull(t *testing.T) {
	v, err := Interests{}.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != nil {
		t.Fatalf("Value of zero Interests = %v, want nil", v)
	}
}

func TestInterestsValueAllEventsRoundTrip(t *testing.T) {
	v, err := AllEvents().Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	raw, ok := v.([]byte)
	if !ok {
		t.Fatalf("Value type = %T, want []byte", v)
	}

	var out Interests
	if err := out.Scan(raw); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !out.AllEvents {
		t.Fatalf("AllEvents lost in round trip: %s", string(raw))
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
