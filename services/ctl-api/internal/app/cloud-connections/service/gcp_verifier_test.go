package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateGCPNegativeProbe(t *testing.T) {
	tests := map[string]struct {
		stage   gcpNegativeProbeStage
		wantErr bool
	}{
		"denied at STS passes":                 {stage: gcpNegativeProbeDeniedAtSTS},
		"denied at generateAccessToken passes": {stage: gcpNegativeProbeDeniedAtImpersonation},
		"both succeed fails":                   {stage: gcpNegativeProbeSucceeded, wantErr: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.wantErr, evaluateGCPNegativeProbe(test.stage) != nil)
		})
	}
}
