package joberrors

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestLifecycleFailureError(t *testing.T) {
	reasons := []LifecycleFailureReason{
		LifecycleFailureReasonNoActiveRunner,
		LifecycleFailureReasonQueueTimeout,
		LifecycleFailureReasonRunnerUnhealthy,
		LifecycleFailureReasonPickupTimeout,
		LifecycleFailureReasonOverallTimeout,
		LifecycleFailureReasonExecutionTimeout,
		LifecycleFailureReasonAttemptsExhausted,
		LifecycleFailureReasonResultMissing,
	}

	for _, reason := range reasons {
		t.Run(string(reason), func(t *testing.T) {
			err := &LifecycleFailureError{Reason: reason}
			if err.Error() == "" {
				t.Fatal("Error() should not be empty")
			}
			if err.Type() != LifecycleFailureErrorType {
				t.Fatalf("Type() = %q, want %q", err.Type(), LifecycleFailureErrorType)
			}
			if err.Severity() != compositeerrors.SeverityError {
				t.Fatalf("Severity() = %q, want %q", err.Severity(), compositeerrors.SeverityError)
			}
			if sections := err.Sections(); len(sections) != 2 || sections[0].Body == "" || sections[1].Body == "" {
				t.Fatalf("expected populated what-happened and how-to-fix sections, got %#v", sections)
			}

			data, newErr := compositeerrors.New(err)
			if newErr != nil {
				t.Fatalf("unexpected error freezing composite error: %v", newErr)
			}
			if data.Type != LifecycleFailureErrorType {
				t.Fatalf("frozen type = %q, want %q", data.Type, LifecycleFailureErrorType)
			}
		})
	}
}
