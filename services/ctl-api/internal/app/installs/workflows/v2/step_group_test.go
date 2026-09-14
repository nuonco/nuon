package v2

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
)

func TestNextGroupEagerParallel(t *testing.T) {
	sg := newStepGroup(&app.Workflow{})

	sg.nextGroupEager()
	first := sg.currentGroup

	sg.nextGroupEagerParallel()
	second := sg.currentGroup

	sg.nextGroupEager()
	third := sg.currentGroup

	if first.Parallel {
		t.Error("nextGroupEager must not open a parallel group")
	}
	if !first.EagerExecution {
		t.Error("nextGroupEager must mark the group eager")
	}

	// Provision puts the runner service account and stack generation in this
	// group so they run concurrently. It has to be parallel to overlap them and
	// eager so the conductor starts it before generation finishes.
	if !second.Parallel {
		t.Error("nextGroupEagerParallel must open a parallel group")
	}
	if !second.EagerExecution {
		t.Error("nextGroupEagerParallel must mark the group eager")
	}

	if second.GroupIdx == first.GroupIdx || third.GroupIdx == second.GroupIdx {
		t.Fatalf("each call must open a new group, got idxs %d, %d, %d",
			first.GroupIdx, second.GroupIdx, third.GroupIdx)
	}
	if third.Parallel {
		t.Error("a following nextGroupEager must not inherit Parallel")
	}
}

// Provision folds state generation and the runner service account into the
// stack-generation signal. That signal is retryable:false by default, so the
// collapsed step has to opt back in or a transient state-generation failure
// stops being user-retryable.
func TestWithRetryableOverridesSignalDefault(t *testing.T) {
	meta := getSignalStepMetadata(generateinstallstackversion.SignalType, false)
	if meta.retryable {
		t.Fatal("generate-install-stack-version is expected to default to retryable:false")
	}

	step := &app.WorkflowStep{Retryable: meta.retryable}
	WithRetryable(true)(step)

	if !step.Retryable {
		t.Error("WithRetryable(true) must override the signal's default")
	}
}
