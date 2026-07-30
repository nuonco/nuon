// Package awaitcomponenthealthy holds the verified-deploy stabilization gate: it
// waits for a component to hold healthy after its apply before the deploy step
// completes. Added at plan time only when the component opts into block_deploy.
package awaitcomponenthealthy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/poll"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "await-component-healthy"

const (
	// noObservationExtension is the only case the gate runs past its configured
	// window: waiting for a first report when the runner is slow or wedged.
	noObservationExtension = 75 * time.Second
	pollInitialInterval    = 10 * time.Second
	pollMaxInterval        = 20 * time.Second
	pollBackoffFactor      = 1.1
)

type Signal struct {
	InstallID          string `json:"install_id"`
	InstallComponentID string `json:"install_component_id"`
	WorkflowStepID     string `json:"workflow_step_id"`

	v *validator.Validate
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithParams         = (*Signal)(nil)
	_ signal.SignalWithStepContext    = (*Signal)(nil)
	_ signal.SignalWithAutoRetry      = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }

// gateMaxAutoRetries bounds self re-checks before parking for a human, leaving
// most of the global retry budget for the retry button.
const gateMaxAutoRetries = 3

func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return gateMaxAutoRetries }

func (s *Signal) WithParams(params *signal.Params) {
	s.v = params.V
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.InstallComponentID == "" {
		return errors.New("install_component_id is required")
	}

	if _, err := activities.AwaitGetInstallComponentByID(ctx, s.InstallComponentID); err != nil {
		return errors.Wrap(err, "install component not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	l = l.With(
		zap.String("install_component_id", s.InstallComponentID),
		zap.String("workflow_step_id", s.WorkflowStepID),
	)

	installComponent, err := activities.AwaitGetInstallComponentByID(ctx, s.InstallComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get install component")
	}

	// Point the step at the component so the dashboard's detail panel can
	// resolve what is being verified. Best-effort: the gate works without it.
	if s.WorkflowStepID != "" {
		_ = activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   s.InstallComponentID,
			StepTargetType: string(app.WorkflowStepTargetTypeInstallComponents),
		})
	}

	if !isWatchableComponentType(installComponent.Component.Type) {
		return nil
	}

	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureComponentHealth))
	if err != nil {
		return errors.Wrap(err, "unable to check component-health feature")
	}
	if !enabled {
		return nil
	}

	ccc, err := s.componentConfig(ctx, installComponent.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component config")
	}
	if ccc == nil || !ccc.HealthCheckEnabled() || !ccc.HealthBlocksDeploy() {
		return nil
	}

	declaredProbes := make([]string, 0, len(ccc.HealthProbes))
	for _, probe := range ccc.HealthProbes {
		if probe.Name != "" {
			declaredProbes = append(declaredProbes, probe.Name)
		}
	}

	window := ccc.HealthStabilization()
	// Judged off raw runner reports rather than the debounced verdict, so the gate
	// carries no evaluator-cron or debounce latency.
	gateStartedAt := workflow.Now(ctx)
	deadline := gateStartedAt.Add(window).Add(noObservationExtension)

	var lastReport *activities.GateHealthReport
	var sawData bool
	var lastNarration string
	var failMsg string
	var awaitingProbes []string
	pollErr := poll.Poll(ctx, s.v, poll.PollOpts{
		MaxTS:           deadline,
		InitialInterval: pollInitialInterval,
		MaxInterval:     pollMaxInterval,
		BackoffFactor:   pollBackoffFactor,
		Fn: func(ctx workflow.Context) error {
			now := workflow.Now(ctx)
			reports, err := activities.AwaitGetGateHealthReports(ctx, &activities.GetGateHealthReportsRequest{
				InstallID:          s.InstallID,
				InstallComponentID: s.InstallComponentID,
				Since:              gateStartedAt.Add(-time.Second),
			})
			if err != nil {
				return err
			}

			outcome, report := judgeWindow(reports, gateStartedAt, now, window, installComponent.ClusterHealthSeen())
			lastReport = report
			if report != nil {
				sawData = true
			}

			// The check snapshot is best-effort context for the step's detail
			// view; a lookup failure must never affect the gate decision.
			checks, cerr := activities.AwaitGetComponentHealthCheckRows(ctx, &activities.GetComponentHealthCheckRowsRequest{
				InstallID:          s.InstallID,
				InstallComponentID: s.InstallComponentID,
			})
			if cerr != nil {
				checks = nil
			}
			markRemovedCheckRows(checks, declaredProbes)

			// The runner picks probes up a cycle late after a first deploy, so
			// without this an all-healthy window could verify a deploy whose
			// declared checks never ran once.
			awaitingProbes = nil
			if outcome == windowPass && len(declaredProbes) > 0 {
				if missing := missingProbes(declaredProbes, checks, gateStartedAt); len(missing) > 0 {
					awaitingProbes = missing
					s.narrate(ctx, &lastNarration, fmt.Sprintf(
						"window elapsed all-healthy — waiting for declared probes to report (%d of %d): missing %s",
						len(declaredProbes)-len(missing), len(declaredProbes), strings.Join(missing, ", ")), checks)
					return errors.Errorf("declared probes have not reported yet")
				}
			}

			s.narrate(ctx, &lastNarration, windowNarration(outcome, report, gateStartedAt, now, window), checks)

			switch outcome {
			case windowPass:
				return nil
			case windowFailBad:
				failMsg = badReportMessage(report)
				return errors.Wrap(poll.NonRetryableError, failMsg)
			case windowFailState:
				failMsg = fmt.Sprintf("component did not hold healthy for %s: window ended %s", window, describeReport(report))
				return errors.Wrap(poll.NonRetryableError, failMsg)
			case windowNeedData:
				return errors.Errorf("no health reports observed since the apply yet")
			default:
				return errors.Errorf("window still open")
			}
		},
		PostAttemptHook: func(ctx workflow.Context, dur time.Duration) error {
			l.Debug("verified-deploy gate still open, checking again", zap.Duration("next_check_in", dur))
			return nil
		},
	})

	if pollErr == nil {
		return nil
	}
	if errors.Is(pollErr, context.DeadlineExceeded) {
		if !sawData {
			return fmt.Errorf("the runner reported no health observations for %s after the apply (window %s + %s waiting for a report)",
				window+noObservationExtension, window, noObservationExtension)
		}
		if len(awaitingProbes) > 0 {
			return fmt.Errorf("declared health probes never reported after the apply: %s — a passing window must include every configured check",
				strings.Join(awaitingProbes, ", "))
		}
		return fmt.Errorf("component did not hold healthy for %s: window ended %s", window, describeReport(lastReport))
	}
	// The poll sentinel is loop control, not user-facing text — return the
	// clean message instead of "...: non-retryable".
	if failMsg != "" && errors.Is(pollErr, poll.NonRetryableError) {
		return errors.New(failMsg)
	}
	return pollErr
}

func isWatchableComponentType(t app.ComponentType) bool {
	return t == app.ComponentTypeHelmChart || t == app.ComponentTypeKubernetesManifest
}

// componentConfig resolves the component's current config, pin first with the
// latest-configs view as fallback. ccc rows are deltas, so the pin alone returns
// nil after any unrelated sync and silently turns the gate off.
func (s *Signal) componentConfig(ctx workflow.Context, componentID string) (*app.ComponentConfigConnection, error) {
	return activities.AwaitGetCurrentComponentConfig(ctx, &activities.GetCurrentComponentConfigRequest{
		InstallID:   s.InstallID,
		ComponentID: componentID,
	})
}

// badReportMessage names the failing observation and what it said.
func badReportMessage(r *activities.GateHealthReport) string {
	if r == nil {
		return "a health observation inside the stabilization window was failing"
	}
	msg := fmt.Sprintf("component is %s", r.Health)
	if r.RootKind != "" {
		msg += ": " + r.RootKind + " " + r.RootName
	}
	if r.Message != "" {
		msg += ": " + r.Message
	}
	return msg
}

func describeReport(r *activities.GateHealthReport) string {
	if r == nil {
		return "with no health observations"
	}
	desc := r.Health
	if r.RootKind != "" {
		desc += " (" + r.RootKind + " " + r.RootName
		if r.Message != "" {
			desc += ": " + r.Message
		}
		desc += ")"
	}
	return desc
}

// narrate writes the gate's progress onto its workflow step, so it isn't a silent
// step that eventually flips. Best-effort — a failure here must not fail the gate.
func (s *Signal) narrate(ctx workflow.Context, lastNarration *string, narration string, checks []activities.ComponentHealthCheckRow) {
	if s.WorkflowStepID == "" || narration == "" {
		return
	}
	key := narration
	for _, c := range checks {
		key += "|" + c.Kind + "/" + c.Name + "=" + c.Health
	}
	if key == *lastNarration {
		return
	}
	*lastNarration = key

	metadata := map[string]any{}
	if len(checks) > 0 {
		// Explicit lowercase keys: this crosses the Temporal payload converter
		// into JSONB, and a struct with mismatched tags renders as empty rows.
		rows := make([]map[string]any, 0, len(checks))
		for _, c := range checks {
			row := map[string]any{
				"kind":    c.Kind,
				"name":    c.Name,
				"health":  c.Health,
				"message": c.Message,
			}
			if c.Removed {
				row["removed"] = true
			}
			rows = append(rows, row)
		}
		metadata["checks"] = rows
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowStepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: narration,
			Metadata:               metadata,
		},
	}); err != nil {
		if l, lerr := log.WorkflowLogger(ctx); lerr == nil {
			l.Warn("unable to narrate verify-health step", zap.Error(err))
		}
	}
}

// windowNarration renders one poll reading as a human sentence with the real
// remaining budget.
func windowNarration(outcome windowOutcome, report *activities.GateHealthReport, gateStartedAt, now time.Time, window time.Duration) string {
	switch outcome {
	case windowPass:
		return fmt.Sprintf("held healthy for %s — %s", window, describeReport(report))
	case windowFailBad:
		// Written just before the step's error lands, so the failing snapshot
		// is locked in the step history with the checks that caused it.
		return badReportMessage(report)
	case windowFailState:
		return fmt.Sprintf("window ended %s", describeReport(report))
	case windowNeedData:
		return fmt.Sprintf("window elapsed with no reports yet — waiting up to %s for the runner's next report", noObservationExtension)
	}

	remaining := gateStartedAt.Add(window).Sub(now).Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}
	if report == nil {
		return fmt.Sprintf("waiting for the first health report since the apply — %s of the window left", remaining)
	}
	return fmt.Sprintf("healthy so far — %s of the %s window left — %s", remaining, window, describeReport(report))
}
