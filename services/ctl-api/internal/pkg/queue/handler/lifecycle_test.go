package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// stubSignal is a minimal signal implementation for testing.
type stubSignal struct{}

func (s *stubSignal) Type() signal.SignalType           { return "test" }
func (s *stubSignal) Validate(_ workflow.Context) error { return nil }
func (s *stubSignal) Execute(_ workflow.Context) error  { return nil }

// stubSignalWithLogStream embeds stubSignal and implements SignalWithLogStream.
type stubSignalWithLogStream struct {
	stubSignal
	id string
}

func (s *stubSignalWithLogStream) LogStreamID() string { return s.id }

func TestLogStreamMetadata(t *testing.T) {
	tests := []struct {
		name string
		sig  signal.Signal
		want map[string]any
	}{
		{
			name: "signal without log stream returns nil",
			sig:  &stubSignal{},
			want: nil,
		},
		{
			name: "signal with empty log stream ID returns nil",
			sig:  &stubSignalWithLogStream{id: ""},
			want: nil,
		},
		{
			name: "signal with log stream ID returns metadata",
			sig:  &stubSignalWithLogStream{id: "ls-123"},
			want: map[string]any{"log_stream_id": "ls-123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := logStreamMetadata(tt.sig)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildOutcome(t *testing.T) {
	dur := 5 * time.Second

	tests := []struct {
		name       string
		sig        signal.Signal
		err        error
		wantStatus signal.SignalStatus
		wantErr    string
		wantMeta   map[string]any
	}{
		{
			name:       "success without log stream",
			sig:        &stubSignal{},
			err:        nil,
			wantStatus: signal.SignalStatusSuccess,
			wantErr:    "",
			wantMeta:   nil,
		},
		{
			name:       "error without log stream",
			sig:        &stubSignal{},
			err:        errors.New("boom"),
			wantStatus: signal.SignalStatusError,
			wantErr:    "boom",
			wantMeta:   nil,
		},
		{
			name:       "success with log stream",
			sig:        &stubSignalWithLogStream{id: "ls-abc"},
			err:        nil,
			wantStatus: signal.SignalStatusSuccess,
			wantErr:    "",
			wantMeta:   map[string]any{"log_stream_id": "ls-abc"},
		},
		{
			name:       "error with log stream",
			sig:        &stubSignalWithLogStream{id: "ls-abc"},
			err:        errors.New("fail"),
			wantStatus: signal.SignalStatusError,
			wantErr:    "fail",
			wantMeta:   map[string]any{"log_stream_id": "ls-abc"},
		},
		{
			name:       "success with empty log stream ID",
			sig:        &stubSignalWithLogStream{id: ""},
			err:        nil,
			wantStatus: signal.SignalStatusSuccess,
			wantErr:    "",
			wantMeta:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outcome := buildOutcome(tt.sig, tt.err, dur)
			assert.Equal(t, tt.wantStatus, outcome.Status)
			assert.Equal(t, tt.wantErr, outcome.ErrMessage)
			assert.Equal(t, dur, outcome.Duration)
			assert.Equal(t, tt.wantMeta, outcome.Metadata)
		})
	}
}
