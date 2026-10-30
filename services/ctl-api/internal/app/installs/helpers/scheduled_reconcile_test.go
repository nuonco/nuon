package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateScheduledReconcileTime(t *testing.T) {
	tests := map[string]struct {
		at      time.Time
		wantErr error
	}{
		"too soon": {
			at:      time.Now().Add(30 * time.Second),
			wantErr: ErrScheduledReconcileTimeTooSoon,
		},
		"exactly at the minimum": {
			at:      time.Now().Add(MinScheduledReconcileLeadTime + time.Minute),
			wantErr: nil,
		},
		"within window": {
			at:      time.Now().Add(24 * time.Hour),
			wantErr: nil,
		},
		"too far": {
			at:      time.Now().Add(MaxScheduledReconcileLeadTime + 24*time.Hour),
			wantErr: ErrScheduledReconcileTimeTooFar,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateScheduledReconcileTime(tt.at)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
