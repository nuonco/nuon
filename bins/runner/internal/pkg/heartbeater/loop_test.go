package heartbeater

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestShouldRestartForHeartBeatError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "inactive process conflict",
			err: &operations.CreateRunnerHeartBeatConflict{Payload: &models.StderrErrResponse{
				Description: inactiveRunnerProcessConflict,
			}},
			want: true,
		},
		{
			name: "other conflict",
			err: &operations.CreateRunnerHeartBeatConflict{Payload: &models.StderrErrResponse{
				Description: "other conflict",
			}},
		},
		{
			name: "transient error",
			err:  errors.New("temporary network failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldRestartForHeartBeatError(tt.err))
		})
	}
}

type shutdownRecorder struct {
	calls int
}

func (s *shutdownRecorder) Shutdown(...fx.ShutdownOption) error {
	s.calls++
	return nil
}

func TestHandleHeartBeatErrorRequestsRestartOnlyForInactiveProcess(t *testing.T) {
	shutdowner := &shutdownRecorder{}
	heartBeater := &HeartBeater{l: zap.NewNop(), shutdowner: shutdowner}
	inactive := &operations.CreateRunnerHeartBeatConflict{Payload: &models.StderrErrResponse{
		Description: inactiveRunnerProcessConflict,
	}}

	assert.True(t, heartBeater.handleHeartBeatError(inactive))
	assert.Equal(t, 1, shutdowner.calls)
	assert.False(t, heartBeater.handleHeartBeatError(errors.New("temporary network failure")))
	assert.Equal(t, 1, shutdowner.calls)
}
