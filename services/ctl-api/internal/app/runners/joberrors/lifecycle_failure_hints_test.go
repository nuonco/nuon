package joberrors

import "testing"

func TestLifecycleFailureHints(t *testing.T) {
	if !(&LifecycleFailureError{Reason: LifecycleFailureReasonRunnerDisabled}).Hints().Terminal() {
		t.Error("a disabled runner should be a terminal failure")
	}

	for _, reason := range []LifecycleFailureReason{
		LifecycleFailureReasonNoActiveRunner,
		LifecycleFailureReasonQueueTimeout,
		LifecycleFailureReasonRunnerUnhealthy,
		LifecycleFailureReasonPickupTimeout,
	} {
		if (&LifecycleFailureError{Reason: reason}).Hints().Terminal() {
			t.Errorf("%s should stay retryable", reason)
		}
	}
}
