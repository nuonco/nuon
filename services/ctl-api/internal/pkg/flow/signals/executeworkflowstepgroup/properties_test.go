package executeworkflowstepgroup

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"go.temporal.io/sdk/workflow"
	"hegel.dev/go/hegel"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

// ============================================================================
// Property: nextExecutableStep always returns the first executable step
// ============================================================================

// allStatuses is every step status the system can produce.
var allStatuses = []app.Status{
	app.StatusPending,
	app.StatusNotAttempted,
	app.StatusQueued,
	app.StatusSuccess,
	app.StatusError,
	app.StatusInProgress,
	app.StatusDiscarded,
	app.StatusCancelled,
	app.StatusAutoSkipped,
	app.StatusUserSkipped,
	app.WorkflowStepApprovalStatusApproved,
	app.WorkflowStepNoDrift,
	app.WorkflowStepDrifted,
}

// executableStatuses is the set nextExecutableStep considers "executable".
var executableStatuses = map[app.Status]bool{
	app.StatusPending:      true,
	app.StatusNotAttempted: true,
	app.StatusQueued:       true,
}

func TestNextExecutableStepProperties(t *testing.T) {
	s := &Signal{}

	t.Run("always picks the first executable step", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 15))
		steps := make([]app.WorkflowStep, n)
		for i := range steps {
			status := hegel.Draw(ht, hegel.SampledFrom(allStatuses))
			steps[i] = app.WorkflowStep{
				ID:     fmt.Sprintf("step-%d", i),
				Idx:    i * 100,
				Status: app.CompositeStatus{Status: status},
			}
		}

		picked, found := s.nextExecutableStep(steps)

		// Find expected first executable step independently.
		expectedIdx := -1
		for j, step := range steps {
			if executableStatuses[step.Status.Status] {
				expectedIdx = j
				break
			}
		}

		if expectedIdx == -1 {
			if found {
				ht.Fatalf("no executable step exists but found=true, picked %s", picked.ID)
			}
		} else {
			if !found {
				ht.Fatalf("executable step exists at index %d but found=false", expectedIdx)
			}
			if picked.ID != steps[expectedIdx].ID {
				ht.Fatalf("picked %s but expected first executable at index %d (%s)",
					picked.ID, expectedIdx, steps[expectedIdx].ID)
			}
		}
	}, hegel.WithTestCases(500)))

	t.Run("terminal steps are never picked", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 15))
		steps := make([]app.WorkflowStep, n)
		for i := range steps {
			status := hegel.Draw(ht, hegel.SampledFrom(allStatuses))
			steps[i] = app.WorkflowStep{
				ID:     fmt.Sprintf("step-%d", i),
				Idx:    i * 100,
				Status: app.CompositeStatus{Status: status},
			}
		}

		picked, found := s.nextExecutableStep(steps)
		if found {
			if !executableStatuses[picked.Status.Status] {
				ht.Fatalf("picked step %s with non-executable status %q",
					picked.ID, picked.Status.Status)
			}
		}
	}, hegel.WithTestCases(500)))
}

// ============================================================================
// Property: nextExecutableStep respects Idx ordering
// ============================================================================

func TestNextExecutableStep_PicksLowestIdx(t *testing.T) {
	s := &Signal{}

	t.Run("picked step has lowest Idx among executables", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 15))
		steps := make([]app.WorkflowStep, n)
		for i := range steps {
			status := hegel.Draw(ht, hegel.SampledFrom(allStatuses))
			idx := hegel.Draw(ht, hegel.Integers(0, 1000))
			steps[i] = app.WorkflowStep{
				ID:     fmt.Sprintf("step-%d", i),
				Idx:    idx,
				Status: app.CompositeStatus{Status: status},
			}
		}

		// Sort by Idx (simulating DB ORDER BY idx ASC).
		sort.Slice(steps, func(a, b int) bool {
			return steps[a].Idx < steps[b].Idx
		})

		picked, found := s.nextExecutableStep(steps)

		// Find the executable step with the lowest Idx independently.
		lowestIdx := -1
		lowestID := ""
		for _, step := range steps {
			if executableStatuses[step.Status.Status] {
				if lowestIdx == -1 || step.Idx < lowestIdx {
					lowestIdx = step.Idx
					lowestID = step.ID
				}
			}
		}

		if lowestIdx == -1 {
			if found {
				ht.Fatalf("no executable step exists but found=true")
			}
		} else {
			if !found {
				ht.Fatalf("executable step exists (Idx=%d) but found=false", lowestIdx)
			}
			if picked.Idx != lowestIdx {
				ht.Fatalf("picked step %s (Idx=%d) but lowest executable Idx is %d (%s)",
					picked.ID, picked.Idx, lowestIdx, lowestID)
			}
		}
	}, hegel.WithTestCases(500)))
}

// ============================================================================
// Property: groupMaxRetriesForSteps is the min across all signals
// ============================================================================

type testMaxRetriesSignal struct {
	maxRetries int
}

func (s *testMaxRetriesSignal) Type() signal.SignalType         { return "test-max-retries" }
func (s *testMaxRetriesSignal) Validate(workflow.Context) error { return nil }
func (s *testMaxRetriesSignal) Execute(workflow.Context) error  { return nil }
func (s *testMaxRetriesSignal) MaxRetries() int                 { return s.maxRetries }
func (s *testMaxRetriesSignal) SleepAfter() time.Duration       { return 0 }

type testNoMaxRetriesSignal struct{}

func (s *testNoMaxRetriesSignal) Type() signal.SignalType         { return "test-no-max" }
func (s *testNoMaxRetriesSignal) Validate(workflow.Context) error { return nil }
func (s *testNoMaxRetriesSignal) Execute(workflow.Context) error  { return nil }
func (s *testNoMaxRetriesSignal) SleepAfter() time.Duration       { return 0 }

type testCloneMaxRetriesSignal struct {
	maxRetries      int
	cloneMaxRetries int
}

func (s *testCloneMaxRetriesSignal) Type() signal.SignalType         { return "test-clone-max" }
func (s *testCloneMaxRetriesSignal) Validate(workflow.Context) error { return nil }
func (s *testCloneMaxRetriesSignal) Execute(workflow.Context) error  { return nil }
func (s *testCloneMaxRetriesSignal) MaxRetries() int                 { return s.maxRetries }
func (s *testCloneMaxRetriesSignal) SleepAfter() time.Duration       { return 0 }

func (s *testCloneMaxRetriesSignal) CloneSteps(originalStepName string) []signal.CloneStepDef {
	return []signal.CloneStepDef{
		{Name: originalStepName + "-plan", Signal: &testMaxRetriesSignal{maxRetries: s.cloneMaxRetries}},
		{Name: originalStepName + "-apply", Signal: &testMaxRetriesSignal{maxRetries: s.maxRetries}},
	}
}

func TestGroupMaxRetriesForStepsProperties(t *testing.T) {
	t.Run("result equals min of all MaxRetries values", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 6))
		steps := make([]app.WorkflowStep, n)
		expectedMin := signal.DefaultMaxRetries

		for j := range steps {
			mr := hegel.Draw(ht, hegel.Integers(1, 20))
			if mr < expectedMin {
				expectedMin = mr
			}
			steps[j] = app.WorkflowStep{
				ID:   fmt.Sprintf("step-%d", j),
				Name: fmt.Sprintf("step-%d", j),
				QueueSignal: &signaldb.SignalData{
					Signal: &testMaxRetriesSignal{maxRetries: mr},
				},
			}
		}

		result := groupMaxRetriesForSteps(steps)
		if result != expectedMin {
			ht.Fatalf("got %d, want min=%d", result, expectedMin)
		}
	}, hegel.WithTestCases(500)))

	t.Run("adding a stricter step can only decrease or preserve", hegel.Case(func(ht *hegel.T) {
		n := hegel.Draw(ht, hegel.Integers(1, 5))
		steps := make([]app.WorkflowStep, n)
		for j := range steps {
			mr := hegel.Draw(ht, hegel.Integers(5, 15))
			steps[j] = app.WorkflowStep{
				ID:   fmt.Sprintf("step-%d", j),
				Name: fmt.Sprintf("step-%d", j),
				QueueSignal: &signaldb.SignalData{
					Signal: &testMaxRetriesSignal{maxRetries: mr},
				},
			}
		}

		baseline := groupMaxRetriesForSteps(steps)

		stricter := hegel.Draw(ht, hegel.Integers(1, 4))
		steps = append(steps, app.WorkflowStep{
			ID:   "stricter",
			Name: "stricter",
			QueueSignal: &signaldb.SignalData{
				Signal: &testMaxRetriesSignal{maxRetries: stricter},
			},
		})

		withStricter := groupMaxRetriesForSteps(steps)
		if withStricter > baseline {
			ht.Fatalf("adding stricter step (max=%d) increased group max: %d -> %d",
				stricter, baseline, withStricter)
		}
	}, hegel.WithTestCases(300)))

	t.Run("clone step signals participate in min", func(t *testing.T) {
		// Primary has maxRetries=10, but clone step has maxRetries=3.
		steps := []app.WorkflowStep{{
			ID:   "step-1",
			Name: "step-1",
			QueueSignal: &signaldb.SignalData{
				Signal: &testCloneMaxRetriesSignal{maxRetries: 10, cloneMaxRetries: 3},
			},
		}}

		result := groupMaxRetriesForSteps(steps)
		if result != 3 {
			t.Fatalf("got %d, want 3 (clone step min)", result)
		}
	})

	t.Run("nil signals are skipped", func(t *testing.T) {
		steps := []app.WorkflowStep{
			{ID: "a", QueueSignal: nil},
			{ID: "b", QueueSignal: &signaldb.SignalData{Signal: nil}},
			{ID: "c", QueueSignal: &signaldb.SignalData{Signal: &testMaxRetriesSignal{maxRetries: 5}}},
		}
		result := groupMaxRetriesForSteps(steps)
		if result != 5 {
			t.Fatalf("got %d, want 5", result)
		}
	})

	t.Run("empty steps returns default", func(t *testing.T) {
		result := groupMaxRetriesForSteps(nil)
		if result != signal.DefaultMaxRetries {
			t.Fatalf("got %d, want %d", result, signal.DefaultMaxRetries)
		}
	})

	t.Run("signals without MaxRetries use default", func(t *testing.T) {
		steps := []app.WorkflowStep{{
			ID:          "a",
			Name:        "a",
			QueueSignal: &signaldb.SignalData{Signal: &testNoMaxRetriesSignal{}},
		}}
		result := groupMaxRetriesForSteps(steps)
		if result != signal.DefaultMaxRetries {
			t.Fatalf("got %d, want %d", result, signal.DefaultMaxRetries)
		}
	})
}
