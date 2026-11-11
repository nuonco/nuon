package handler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
)

func sdkUpdateRejection(msg string) error {
	fc := temporal.GetDefaultFailureConverter()
	return fc.FailureToError(fc.ErrorToFailure(errors.New(msg)))
}

func TestIsUnknownUpdate(t *testing.T) {
	rejection := sdkUpdateRejection("unknown update ready. KnownUpdates=[]")

	tests := map[string]struct {
		err        error
		updateName string
		want       bool
	}{
		"sdk rejection": {
			err:        rejection,
			updateName: ReadyHandlerName,
			want:       true,
		},
		"sdk rejection wrapped with pkg errors": {
			err:        pkgerrors.Wrap(rejection, "unable get response"),
			updateName: ReadyHandlerName,
			want:       true,
		},
		"sdk rejection wrapped with fmt": {
			err:        fmt.Errorf("unable get response: %w", rejection),
			updateName: ReadyHandlerName,
			want:       true,
		},
		"plain string error": {
			err:        errors.New("unknown update ready. KnownUpdates=[]"),
			updateName: ReadyHandlerName,
			want:       false,
		},
		"other update name": {
			err:        rejection,
			updateName: "execute",
			want:       false,
		},
		"update name prefix": {
			err:        sdkUpdateRejection("unknown update ready-extra. KnownUpdates=[]"),
			updateName: ReadyHandlerName,
			want:       false,
		},
		"nil": {
			err:        nil,
			updateName: ReadyHandlerName,
			want:       false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var appErr *temporal.ApplicationError
			if tc.want {
				require.ErrorAs(t, tc.err, &appErr)
			}
			assert.Equal(t, tc.want, IsUnknownUpdate(tc.err, tc.updateName))
		})
	}
}

func TestRetryUpdate(t *testing.T) {
	tests := map[string]struct {
		timeout   time.Duration
		cancelOn  int
		blockOn   int
		succeedOn int
		retryable bool
		minCalls  int
		maxCalls  int
		wantErr   error
	}{
		"deadline inside the first delay returns without sleeping": {
			timeout:   300 * time.Millisecond,
			retryable: true,
			minCalls:  1,
			maxCalls:  1,
			wantErr:   context.DeadlineExceeded,
		},
		"deadline stops retrying before the delay would cross it": {
			timeout:   1300 * time.Millisecond,
			retryable: true,
			minCalls:  2,
			maxCalls:  3,
			wantErr:   context.DeadlineExceeded,
		},
		"blocking attempt returns before the outer deadline": {
			timeout:   1500 * time.Millisecond,
			blockOn:   1,
			retryable: true,
			minCalls:  1,
			maxCalls:  1,
			wantErr:   context.DeadlineExceeded,
		},
		"canceled context returns the last error": {
			cancelOn:  1,
			retryable: true,
			minCalls:  1,
			maxCalls:  1,
			wantErr:   context.Canceled,
		},
		"late success on attempt 2": {
			timeout:   10 * time.Second,
			succeedOn: 2,
			retryable: true,
			minCalls:  2,
			maxCalls:  2,
		},
		"non-retryable error returns immediately": {
			timeout:  10 * time.Second,
			minCalls: 1,
			maxCalls: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc
			if tc.timeout > 0 {
				ctx, cancel = context.WithTimeout(context.Background(), tc.timeout)
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()

			calls := 0
			start := time.Now()
			err := RetryUpdate(ctx, ReadyHandlerName, func(error) bool {
				return tc.retryable
			}, func(attemptCtx context.Context) error {
				calls++
				if calls == tc.succeedOn {
					return nil
				}
				if calls == tc.blockOn {
					<-attemptCtx.Done()
					return fmt.Errorf("attempt %d failed: %w", calls, attemptCtx.Err())
				}
				if calls == tc.cancelOn {
					cancel()
				}
				return fmt.Errorf("attempt %d failed", calls)
			})
			elapsed := time.Since(start)

			assert.GreaterOrEqual(t, calls, tc.minCalls)
			assert.LessOrEqual(t, calls, tc.maxCalls)
			if tc.succeedOn > 0 {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.ErrorContains(t, err, fmt.Sprintf("attempt %d failed", calls))
			assert.NotContains(t, err.Error(), "\n")
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			}
			if tc.timeout > 0 {
				assert.Less(t, elapsed, tc.timeout)
			}
			if tc.blockOn > 0 {
				assert.Less(t, elapsed, tc.timeout-updateWithStartDeadlineGap/2)
			}
		})
	}
}
