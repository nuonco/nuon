package executeworkflowstep

import (
	"testing"

	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

type failureCase struct {
	skippable      bool
	retryGroup     bool
	skipAutoRetry  bool
	retryIndex     int
	maxRetries     int
	maxAutoRetries int
}

func drawFailureCase(ht *hegel.T) failureCase {
	return failureCase{
		skippable:      hegel.Draw(ht, hegel.Booleans()),
		retryGroup:     hegel.Draw(ht, hegel.Booleans()),
		skipAutoRetry:  hegel.Draw(ht, hegel.Booleans()),
		retryIndex:     hegel.Draw(ht, hegel.Integers(0, 12)),
		maxRetries:     hegel.Draw(ht, hegel.Integers(0, 12)),
		maxAutoRetries: hegel.Draw(ht, hegel.Integers(0, 12)),
	}
}

func (c failureCase) resolve() directive.Step {
	return resolveFailureDirective(c.skippable, c.retryGroup, c.skipAutoRetry,
		c.retryIndex, c.maxRetries, c.maxAutoRetries)
}

// TestFailedNonSkippableStepNeverContinues asserts the core failure property:
// across the whole retry-budget space, a failed step only ever yields
// StepContinue — the one directive that lets the group advance past it — when
// the step is explicitly skippable. A non-skippable failure always retries,
// parks, or stops; it never carries on.
func TestFailedNonSkippableStepNeverContinues(t *testing.T) {
	t.Run("continue requires skippable", hegel.Case(func(ht *hegel.T) {
		c := drawFailureCase(ht)
		d := c.resolve()

		if d == directive.StepContinue && !c.skippable {
			ht.Fatalf("non-skippable failure continued: %+v", c)
		}
	}, hegel.WithTestCases(300)))
}

// TestExhaustedNonSkippableFailureAlwaysStops asserts that once the global
// retry ceiling is hit, a non-skippable failure resolves to StepStop — the
// directive the group turns into a halt — and nothing else.
func TestExhaustedNonSkippableFailureAlwaysStops(t *testing.T) {
	t.Run("exhausted budget stops", hegel.Case(func(ht *hegel.T) {
		c := drawFailureCase(ht)
		c.skippable = false
		if c.retryIndex+1 <= c.maxRetries {
			return
		}

		if d := c.resolve(); d != directive.StepStop {
			ht.Fatalf("exhausted non-skippable failure resolved to %q, not stop: %+v", d, c)
		}
	}, hegel.WithTestCases(300)))
}

// TestFailureDirectiveIsAlwaysHandled asserts the decision is total over the
// directives the group loop understands: no failure input can produce a
// directive outside {continue, stop, await-retry, retry, retry-group}, which
// would trip the group's fail-closed default.
func TestFailureDirectiveIsAlwaysHandled(t *testing.T) {
	t.Run("decision is total", hegel.Case(func(ht *hegel.T) {
		c := drawFailureCase(ht)

		switch d := c.resolve(); d {
		case directive.StepContinue, directive.StepStop, directive.StepAwaitRetry,
			directive.StepRetry, directive.StepRetryGroup:
		default:
			ht.Fatalf("failure resolved to unexpected directive %q: %+v", d, c)
		}
	}, hegel.WithTestCases(300)))
}

// TestRetryBudgetRemainingNeverStops asserts a failure with retry budget
// remaining is never terminal: it auto-retries (honoring the retry-group
// capability) or parks for manual retry, preserving the user's ability to
// retry up to maxRetries.
func TestRetryBudgetRemainingNeverStops(t *testing.T) {
	t.Run("budget remaining retries or parks", hegel.Case(func(ht *hegel.T) {
		c := drawFailureCase(ht)
		c.skippable = false
		if c.retryIndex+1 > c.maxRetries {
			return
		}

		switch d := c.resolve(); d {
		case directive.StepAwaitRetry:
		case directive.StepRetry:
			if c.retryGroup {
				ht.Fatalf("retry-group signal auto-retried as individual step: %+v", c)
			}
		case directive.StepRetryGroup:
			if !c.retryGroup {
				ht.Fatalf("non-retry-group signal retried as group: %+v", c)
			}
		default:
			ht.Fatalf("failure with budget remaining resolved to %q: %+v", d, c)
		}
	}, hegel.WithTestCases(300)))
}
