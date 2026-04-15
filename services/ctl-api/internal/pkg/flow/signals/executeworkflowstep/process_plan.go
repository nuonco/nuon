package executeworkflowstep

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// planCheck represents a single plan evaluation step that may short-circuit
// the approval flow by returning done=true.
type planCheck struct {
	name      string
	shouldRun func() bool
	run       func() (done bool, err error)
}

// processPlan runs all plan-related checks for an approval step.
// It is called from Execute() after the inner signal completes successfully
// and the step is an approval type. If all checks pass without short-circuiting,
// it proceeds to await user approval.
//
// Each check is implemented in its own file:
//   - process_plan_noop.go    — noop plan detection and auto-skip
//   - process_plan_policy.go  — policy evaluation and violation handling
//   - process_plan_only.go    — plan-only mode auto-approval
func (s *Signal) processPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	l, _ := log.WorkflowLogger(ctx)

	l.Debug("looking up approval contents",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCheckPlan,
			StatusHumanDescription: "checking plan for changes",
			Metadata: map[string]any{
				"status": "checking plan for changes",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step status as checking plan")
	}

	sig := stepSignal(step)
	var noopPlan bool

	checks := []planCheck{
		s.noopCheck(ctx, l, step, flw, sig, &noopPlan),
		s.policyCheck(ctx, l, step, flw, sig),
		s.planOnlyCheck(ctx, step, flw, &noopPlan),
	}

	for _, check := range checks {
		if !check.shouldRun() {
			continue
		}
		done, err := check.run()
		if err != nil {
			if statusErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status: app.StatusError,
					Metadata: map[string]any{
						"reason": fmt.Sprintf("Step failed during %s.", check.name),
					},
					StatusHumanDescription: "Step failed",
				},
			}); statusErr != nil {
				return errors.Wrap(statusErr, "unable to mark step as error")
			}
			return err
		}
		if done {
			return nil
		}
	}

	// All checks passed: proceed to await user approval
	return s.awaitAndHandleApproval(ctx, step, flw)
}

// noopCheck builds the noop plan detection check.
func (s *Signal) noopCheck(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, sig signal.Signal, noopPlan *bool) planCheck {
	return planCheck{
		name: "noop-check",
		shouldRun: func() bool {
			nc, ok := sig.(signal.SignalWithNoOpCheck)
			return ok && nc.IsNoOpCheckable()
		},
		run: func() (bool, error) {
			return s.runNoopCheck(ctx, l, step, flw, noopPlan)
		},
	}
}

// policyCheck builds the policy evaluation check.
func (s *Signal) policyCheck(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, sig signal.Signal) planCheck {
	return planCheck{
		name: "policy-evaluation",
		shouldRun: func() bool {
			pe, ok := sig.(signal.SignalWithPolicyEvaluation)
			return ok && pe.RequiresPolicyEvaluation()
		},
		run: func() (bool, error) {
			return s.runPolicyCheck(ctx, l, step, flw)
		},
	}
}

// planOnlyCheck builds the plan-only auto-approval check.
func (s *Signal) planOnlyCheck(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, noopPlan *bool) planCheck {
	return planCheck{
		name: "plan-only-auto-approval",
		shouldRun: func() bool {
			return flw.PlanOnly
		},
		run: func() (bool, error) {
			return s.runPlanOnlyCheck(ctx, step, *noopPlan)
		},
	}
}
