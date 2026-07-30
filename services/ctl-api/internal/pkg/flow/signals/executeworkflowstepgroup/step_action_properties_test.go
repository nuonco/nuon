package executeworkflowstepgroup

import (
	"testing"

	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

var knownStepDirectives = []directive.Step{
	directive.StepContinue,
	directive.StepStop,
	directive.StepRetry,
	directive.StepRetryGroup,
	directive.StepSkipGroup,
	directive.StepAwaitApproval,
	directive.StepAwaitRetry,
}

func isKnownStepDirective(d directive.Step) bool {
	for _, k := range knownStepDirectives {
		if d == k {
			return true
		}
	}
	return false
}

func drawStepDirective(ht *hegel.T) directive.Step {
	known := make([]string, len(knownStepDirectives))
	for i, d := range knownStepDirectives {
		known[i] = string(d)
	}
	return directive.Step(hegel.Draw(ht, hegel.OneOf(
		hegel.SampledFrom(known),
		hegel.Text(0, 20),
	)))
}

// TestOnlyExplicitSuccessOrRetryDispatchesAnotherStep asserts the group's
// core dispatch property: across all directives (known and arbitrary) and
// flag combinations, the sequential loop only dispatches another step for an
// explicit success (StepContinue), an explicit step retry (StepRetry), or the
// legacy non-resident await-retry passthrough. Everything else — including
// StepStop written for cancelled steps and non-skippable terminal failures —
// halts the group before any further dispatch.
func TestOnlyExplicitSuccessOrRetryDispatchesAnotherStep(t *testing.T) {
	t.Run("dispatching actions require explicit directives", hegel.Case(func(ht *hegel.T) {
		d := drawStepDirective(ht)
		resident := hegel.Draw(ht, hegel.Booleans())
		manual := hegel.Draw(ht, hegel.Booleans())

		action := resolveStepAction(d, resident, manual)

		switch action {
		case actionAdvance:
			legacyAwaitRetry := d == directive.StepAwaitRetry && !resident
			if d != directive.StepContinue && !legacyAwaitRetry {
				ht.Fatalf("directive %q (resident=%v manual=%v) advanced to the next step", d, resident, manual)
			}
		case actionRetryStep:
			if d != directive.StepRetry {
				ht.Fatalf("directive %q (resident=%v manual=%v) re-dispatched as step retry", d, resident, manual)
			}
		}
	}, hegel.WithTestCases(200)))
}

// TestStoppedStepNeverDispatchesAnotherStep pins the cancellation and failure
// halt: handleStepCancelled and non-skippable exhausted failures both write
// StepStop, which must always resolve to a group stop regardless of flags.
func TestStoppedStepNeverDispatchesAnotherStep(t *testing.T) {
	t.Run("stop always halts the group", hegel.Case(func(ht *hegel.T) {
		resident := hegel.Draw(ht, hegel.Booleans())
		manual := hegel.Draw(ht, hegel.Booleans())

		if action := resolveStepAction(directive.StepStop, resident, manual); action != actionStopGroup {
			ht.Fatalf("StepStop (resident=%v manual=%v) resolved to %v, not group stop", resident, manual, action)
		}
	}, hegel.WithTestCases(50)))
}

// TestUnknownDirectiveFailsClosed asserts that a directive the group does not
// recognize — an empty or corrupt ResultDirective, or a value written by a
// newer deploy — stops the group instead of silently advancing.
func TestUnknownDirectiveFailsClosed(t *testing.T) {
	t.Run("unrecognized directives stop the group", hegel.Case(func(ht *hegel.T) {
		d := directive.Step(hegel.Draw(ht, hegel.Text(0, 30)))
		if isKnownStepDirective(d) {
			return
		}
		resident := hegel.Draw(ht, hegel.Booleans())
		manual := hegel.Draw(ht, hegel.Booleans())

		if action := resolveStepAction(d, resident, manual); action != actionStopGroup {
			ht.Fatalf("unknown directive %q (resident=%v manual=%v) resolved to %v, not group stop", d, resident, manual, action)
		}
	}, hegel.WithTestCases(200)))
}

// TestResidentManualRetryAlwaysParks asserts resident flows never clone-retry
// on a manual retry: they park as await-retry so the flow host schedules it.
func TestResidentManualRetryAlwaysParks(t *testing.T) {
	t.Run("resident manual retry parks", hegel.Case(func(ht *hegel.T) {
		d := directive.Step(hegel.Draw(ht, hegel.SampledFrom([]string{
			string(directive.StepRetry),
			string(directive.StepRetryGroup),
			string(directive.StepAwaitRetry),
		})))

		if action := resolveStepAction(d, true, true); action != actionAwaitRetry {
			ht.Fatalf("resident manual %q resolved to %v, not await-retry", d, action)
		}
	}, hegel.WithTestCases(30)))
}
