package executeflow

import (
	"fmt"
	"testing"

	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ============================================================================
// Property: collectGroupIndices preserves first-occurrence order and dedupes
// ============================================================================

func TestCollectGroupIndicesProperties(t *testing.T) {
	t.Run("each group appears exactly once", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 25))
		steps := make([]app.WorkflowStep, n)
		for j := range steps {
			steps[j] = app.WorkflowStep{
				ID:       fmt.Sprintf("step-%d", j),
				Idx:      j,
				GroupIdx: hegel.Draw(ht, hegel.Integers(0, 5)),
			}
		}

		result := collectGroupIndices(steps)

		// Must be deduplicated.
		seen := make(map[int]bool)
		for _, g := range result {
			if seen[g] {
				ht.Fatalf("group %d appears more than once in %v", g, result)
			}
			seen[g] = true
		}

		// Must contain every GroupIdx present in steps.
		expected := make(map[int]bool)
		for _, step := range steps {
			expected[step.GroupIdx] = true
		}
		for g := range expected {
			if !seen[g] {
				ht.Fatalf("group %d present in steps but missing from result %v", g, result)
			}
		}
	}, hegel.WithTestCases(500)))

	t.Run("order matches first occurrence in step list", hegel.Case(func(ht *hegel.T) {
		nGroups := hegel.Draw(ht, hegel.Integers(1, 6))
		groupIdxs := make([]int, nGroups)
		for j := range groupIdxs {
			groupIdxs[j] = hegel.Draw(ht, hegel.Integers(0, 20))
		}
		// Dedupe to get the expected order (first occurrence of each).
		var expectedOrder []int
		seen := make(map[int]bool)
		for _, g := range groupIdxs {
			if !seen[g] {
				seen[g] = true
				expectedOrder = append(expectedOrder, g)
			}
		}

		// Build steps: each group has 1-3 steps, in groupIdxs order.
		var steps []app.WorkflowStep
		idx := 0
		for _, gIdx := range groupIdxs {
			nSteps := hegel.Draw(ht, hegel.Integers(1, 3))
			for s := 0; s < nSteps; s++ {
				steps = append(steps, app.WorkflowStep{
					ID:       fmt.Sprintf("step-%d", idx),
					Idx:      idx,
					GroupIdx: gIdx,
				})
				idx++
			}
		}

		result := collectGroupIndices(steps)

		if len(result) != len(expectedOrder) {
			ht.Fatalf("got %d groups %v, want %d groups %v",
				len(result), result, len(expectedOrder), expectedOrder)
		}
		for i := range result {
			if result[i] != expectedOrder[i] {
				ht.Fatalf("result[%d]=%d, want %d (result=%v, expected=%v)",
					i, result[i], expectedOrder[i], result, expectedOrder)
			}
		}
	}, hegel.WithTestCases(300)))

	t.Run("retry clones do not corrupt order", hegel.Case(func(ht *hegel.T) {
		nGroups := hegel.Draw(ht, hegel.Integers(2, 5))
		var steps []app.WorkflowStep
		idx := 0
		for g := 0; g < nGroups; g++ {
			gIdx := (g + 1) * 100
			nSteps := hegel.Draw(ht, hegel.Integers(1, 3))
			for s := 0; s < nSteps; s++ {
				steps = append(steps, app.WorkflowStep{
					ID:       fmt.Sprintf("step-%d", idx),
					Idx:      idx,
					GroupIdx: gIdx,
				})
				idx++
			}
		}

		baseline := collectGroupIndices(steps)

		// Append retry clones for a random group.
		cloneGroupPos := hegel.Draw(ht, hegel.Integers(0, len(baseline)-1))
		retryGroupIdx := baseline[cloneGroupPos]
		nClones := hegel.Draw(ht, hegel.Integers(1, 3))
		for s := 0; s < nClones; s++ {
			steps = append(steps, app.WorkflowStep{
				ID:       fmt.Sprintf("retry-%d", idx),
				Idx:      idx,
				GroupIdx: retryGroupIdx,
			})
			idx++
		}

		withRetry := collectGroupIndices(steps)

		if len(withRetry) != len(baseline) {
			ht.Fatalf("retry clones changed group count: %v -> %v", baseline, withRetry)
		}
		for i := range baseline {
			if withRetry[i] != baseline[i] {
				ht.Fatalf("retry clones changed group order at [%d]: %v -> %v", i, baseline, withRetry)
			}
		}
	}, hegel.WithTestCases(300)))
}

// ============================================================================
// Property: isGroupParallel depends only on steps in that group
// ============================================================================

func TestIsGroupParallelProperties(t *testing.T) {
	t.Run("other groups cannot affect result", hegel.Case(func(ht *hegel.T) {
		targetGroup := 1
		n := hegel.Draw(ht, hegel.Integers(1, 5))
		steps := make([]app.WorkflowStep, n)
		for j := range steps {
			steps[j] = app.WorkflowStep{
				ID:            fmt.Sprintf("target-%d", j),
				GroupIdx:      targetGroup,
				GroupParallel: false,
			}
		}

		baseResult := isGroupParallel(steps, targetGroup)

		// Add parallel steps in OTHER groups.
		nOther := hegel.Draw(ht, hegel.Integers(1, 5))
		for j := 0; j < nOther; j++ {
			otherGroup := hegel.Draw(ht, hegel.Integers(2, 10))
			steps = append(steps, app.WorkflowStep{
				ID:            fmt.Sprintf("other-%d", j),
				GroupIdx:      otherGroup,
				GroupParallel: true,
			})
		}

		withOthers := isGroupParallel(steps, targetGroup)
		if baseResult != withOthers {
			ht.Fatalf("adding parallel steps in other groups changed result: %v -> %v", baseResult, withOthers)
		}
	}, hegel.WithTestCases(300)))

	t.Run("any parallel step in target group flips true", hegel.Case(func(ht *hegel.T) {
		targetGroup := 1
		n := hegel.Draw(ht, hegel.Integers(2, 8))
		steps := make([]app.WorkflowStep, n)
		for j := range steps {
			steps[j] = app.WorkflowStep{
				ID:            fmt.Sprintf("step-%d", j),
				GroupIdx:      targetGroup,
				GroupParallel: false,
			}
		}

		if isGroupParallel(steps, targetGroup) {
			ht.Fatalf("all GroupParallel=false but isGroupParallel returned true")
		}

		// Set one random step to parallel.
		flipIdx := hegel.Draw(ht, hegel.Integers(0, n-1))
		steps[flipIdx].GroupParallel = true

		if !isGroupParallel(steps, targetGroup) {
			ht.Fatalf("set step[%d].GroupParallel=true but isGroupParallel returned false", flipIdx)
		}
	}, hegel.WithTestCases(300)))
}

// ============================================================================
// Property: isStepTerminal exhaustiveness
// ============================================================================

func TestIsStepTerminalExhaustive(t *testing.T) {
	terminal := []app.Status{
		app.StatusSuccess, app.StatusAutoSkipped, app.StatusUserSkipped,
		app.StatusDiscarded, app.StatusCancelled, app.StatusError,
		app.StatusNotAttempted,
		app.WorkflowStepApprovalStatusApproved, app.WorkflowStepApprovalStatusApprovalDenied,
		app.WorkflowStepNoDrift, app.WorkflowStepDrifted,
	}
	for _, s := range terminal {
		if !isStepTerminal(s) {
			t.Errorf("status %q should be terminal", s)
		}
	}

	nonTerminal := []app.Status{
		app.StatusPending, app.StatusQueued,
		app.StatusInProgress, app.StatusRetrying,
	}
	for _, s := range nonTerminal {
		if isStepTerminal(s) {
			t.Errorf("status %q should not be terminal", s)
		}
	}
}
