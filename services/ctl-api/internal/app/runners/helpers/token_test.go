package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestTokensToKeep(t *testing.T) {
	pending := func(id, tokenID string) app.RunnerProcess {
		return app.RunnerProcess{
			ID:              id,
			TokenID:         tokenID,
			CompositeStatus: app.CompositeStatus{Status: app.StatusPending},
		}
	}
	live := func(id, tokenID string) app.RunnerProcess {
		return app.RunnerProcess{
			ID:              id,
			TokenID:         tokenID,
			CompositeStatus: app.CompositeStatus{Status: app.Status(app.RunnerProcessStatusActive)},
		}
	}

	tests := map[string]struct {
		processes []app.RunnerProcess
		want      []string
		wantErr   string
	}{
		"no live processes": {
			processes: []app.RunnerProcess{pending("proc-pending", "tok-pending")},
			wantErr:   "runner has no live processes",
		},
		"live process has not recorded a token": {
			processes: []app.RunnerProcess{live("proc-live", "")},
			wantErr:   "runner process proc-live has not recorded its token yet",
		},
		"pending process without a token does not block": {
			processes: []app.RunnerProcess{
				live("proc-live", "tok-live"),
				pending("proc-pending", ""),
			},
			want: []string{"tok-live"},
		},
		"pending process token is kept": {
			processes: []app.RunnerProcess{
				live("proc-live", "tok-live"),
				pending("proc-pending", "tok-pending"),
			},
			want: []string{"tok-live", "tok-pending"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tokensToKeep(tt.processes)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, ErrTokenCleanupNotReady)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
