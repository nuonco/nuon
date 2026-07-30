package callback

import (
	"errors"
	"testing"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"hegel.dev/go/hegel"
)

type awaitOutcome struct {
	CarriedOn bool
	Cancelled bool
	Failed    bool
	ErrMsg    string
}

func awaitPropertyWorkflow(ctx workflow.Context) (awaitOutcome, error) {
	ref := New(ctx, "await-property")
	result, err := AwaitWithTimeout(ctx, ref, time.Minute)
	if err != nil {
		// Capture the ApplicationError's message rather than err.Error():
		// the latter appends the error type and would mask an empty message.
		errMsg := err.Error()
		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) {
			errMsg = appErr.Message()
		}
		return awaitOutcome{
			Cancelled: IsCancelled(err),
			Failed:    !IsCancelled(err),
			ErrMsg:    errMsg,
		}, nil
	}
	return awaitOutcome{CarriedOn: result != nil}, nil
}

// TestAwaitNeverCarriesOnWhenCancelled asserts the core completion-callback
// property: a parent awaiting a child signal carries on only when the child
// did not terminate as error or cancelled, and a cancelled child is always
// surfaced as a distinct cancellation outcome — never as success, never as a
// plain failure that would enter retry handling.
func TestAwaitNeverCarriesOnWhenCancelled(t *testing.T) {
	knownStatuses := []string{
		"success",
		"error",
		"cancelled",
		"failed-pending-retry",
		"in-progress",
		"discarded",
		"canceled",
		"CANCELLED",
		"",
	}

	t.Run("await_outcome_matches_terminal_status", hegel.Case(func(ht *hegel.T) {
		status := hegel.Draw(ht, hegel.OneOf(
			hegel.SampledFrom(knownStatuses),
			hegel.Text(0, 20),
		))
		desc := hegel.Draw(ht, hegel.Text(0, 40))

		var suite testsuite.WorkflowTestSuite
		env := suite.NewTestWorkflowEnvironment()
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalName("await-property"), Result{
				Status:            status,
				StatusDescription: desc,
			})
		}, 0)

		env.ExecuteWorkflow(awaitPropertyWorkflow)
		if err := env.GetWorkflowError(); err != nil {
			ht.Fatalf("workflow failed: %v", err)
		}
		var outcome awaitOutcome
		if err := env.GetWorkflowResult(&outcome); err != nil {
			ht.Fatalf("unable to read workflow result: %v", err)
		}

		switch status {
		case "cancelled":
			if outcome.CarriedOn {
				ht.Fatalf("carried on after cancelled callback (desc=%q)", desc)
			}
			if !outcome.Cancelled {
				ht.Fatalf("cancelled callback not classified as cancellation: %+v", outcome)
			}
		case "error":
			if outcome.CarriedOn {
				ht.Fatalf("carried on after error callback (desc=%q)", desc)
			}
			if !outcome.Failed {
				ht.Fatalf("error callback not classified as failure: %+v", outcome)
			}
		default:
			if !outcome.CarriedOn {
				ht.Fatalf("did not carry on for status=%q: %+v", status, outcome)
			}
		}

		if outcome.Cancelled && outcome.Failed {
			ht.Fatalf("outcome classified as both cancelled and failed: %+v", outcome)
		}

		// The error message becomes user-visible failure text on the
		// parent — it must never be empty, even when the sender provided
		// no status description.
		if (outcome.Cancelled || outcome.Failed) && outcome.ErrMsg == "" {
			ht.Fatalf("terminal outcome carries empty error message (status=%q desc=%q): %+v",
				status, desc, outcome)
		}
	}, hegel.WithTestCases(100)))
}
