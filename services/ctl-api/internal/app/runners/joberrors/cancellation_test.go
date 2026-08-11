package joberrors

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestCancellationError(t *testing.T) {
	reasons := []CancellationReason{
		CancellationReasonAPI,
		CancellationReasonAttemptsExhausted,
	}

	for _, reason := range reasons {
		t.Run(string(reason), func(t *testing.T) {
			err := &CancellationError{Reason: reason}
			if err.Error() == "" {
				t.Fatal("Error() should not be empty")
			}
			if err.Type() != CancellationErrorType {
				t.Fatalf("Type() = %q, want %q", err.Type(), CancellationErrorType)
			}
			if err.Severity() != compositeerrors.SeverityWarning {
				t.Fatalf("Severity() = %q, want %q", err.Severity(), compositeerrors.SeverityWarning)
			}
			if sections := err.Sections(); len(sections) != 2 || sections[0].Body == "" || sections[1].Body == "" {
				t.Fatalf("expected populated what-happened and how-to-continue sections, got %#v", sections)
			}

			data, newErr := compositeerrors.New(err)
			if newErr != nil {
				t.Fatalf("unexpected error freezing composite error: %v", newErr)
			}
			if data.Type != CancellationErrorType {
				t.Fatalf("frozen type = %q, want %q", data.Type, CancellationErrorType)
			}
		})
	}
}
