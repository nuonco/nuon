package cronutil

import (
	"testing"
	"time"

	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Emitter IDs with pinned fnv32a values so a hash-algorithm change fails loudly:
//
//	qemtest1: 891724809 (%5=4, %15=9, %60=9)
//	qemtest2: 841391952 (%15=12)
//	qemtest5: 958835285 (%5=0)
//	qemtest7: 925280047 (%15=7)
func TestApplyCronJitter(t *testing.T) {
	tests := []struct {
		name      string
		emitterID string
		schedule  string
		window    time.Duration
		want      string
	}{
		{
			name:      "star minute unchanged",
			emitterID: "qemtest1",
			schedule:  "* * * * *",
			window:    15 * time.Minute,
			want:      "* * * * *",
		},
		{
			name:      "step shifted",
			emitterID: "qemtest7",
			schedule:  "*/15 * * * *",
			window:    15 * time.Minute,
			want:      "7-59/15 * * * *",
		},
		{
			name:      "step shifted other emitter",
			emitterID: "qemtest2",
			schedule:  "*/15 * * * *",
			window:    15 * time.Minute,
			want:      "12-59/15 * * * *",
		},
		{
			name:      "step zero offset unchanged",
			emitterID: "qemtest5",
			schedule:  "*/5 * * * *",
			window:    5 * time.Minute,
			want:      "*/5 * * * *",
		},
		{
			name:      "step offset capped by step",
			emitterID: "qemtest1",
			schedule:  "*/5 * * * *",
			window:    60 * time.Minute,
			want:      "4-59/5 * * * *",
		},
		{
			name:      "single value shifted",
			emitterID: "qemtest1",
			schedule:  "0 * * * *",
			window:    60 * time.Minute,
			want:      "9 * * * *",
		},
		{
			name:      "single value wraps mod 60",
			emitterID: "qemtest1",
			schedule:  "59 * * * *",
			window:    60 * time.Minute,
			want:      "8 * * * *",
		},
		{
			name:      "comma list shifted and sorted",
			emitterID: "qemtest1",
			schedule:  "55,25 * * * *",
			window:    60 * time.Minute,
			want:      "4,34 * * * *",
		},
		{
			name:      "range unchanged",
			emitterID: "qemtest1",
			schedule:  "10-20 * * * *",
			window:    60 * time.Minute,
			want:      "10-20 * * * *",
		},
		{
			name:      "invalid expression unchanged",
			emitterID: "qemtest1",
			schedule:  "not a cron",
			window:    60 * time.Minute,
			want:      "not a cron",
		},
		{
			name:      "descriptor unchanged",
			emitterID: "qemtest1",
			schedule:  "@daily",
			window:    60 * time.Minute,
			want:      "@daily",
		},
		{
			name:      "zero window unchanged",
			emitterID: "qemtest1",
			schedule:  "0 * * * *",
			window:    0,
			want:      "0 * * * *",
		},
		{
			name:      "sub-minute window unchanged",
			emitterID: "qemtest1",
			schedule:  "0 * * * *",
			window:    30 * time.Second,
			want:      "0 * * * *",
		},
		{
			name:      "empty emitter id unchanged",
			emitterID: "",
			schedule:  "0 * * * *",
			window:    60 * time.Minute,
			want:      "0 * * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyCronJitter(tt.emitterID, tt.schedule, tt.window)
			assert.Equal(t, tt.want, got)

			again := ApplyCronJitter(tt.emitterID, tt.schedule, tt.window)
			assert.Equal(t, got, again, "must be deterministic")

			if _, err := cron.ParseStandard(tt.schedule); err == nil {
				_, err := cron.ParseStandard(got)
				require.NoError(t, err, "jittered schedule must remain valid")
			}
		})
	}
}
