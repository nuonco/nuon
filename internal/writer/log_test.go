package writer

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/workers-executors/internal/event"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLog(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        func(*testing.T) []logEventWriterOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: func(t *testing.T) []logEventWriterOption {
				return []logEventWriterOption{
					WithLog(zap.NewNop()),
				}
			},
		},
		"missing validator": {
			v: nil,
			opts: func(t *testing.T) []logEventWriterOption {
				return []logEventWriterOption{
					WithLog(zap.NewNop()),
				}
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing log": {
			v:           v,
			opts:        func(t *testing.T) []logEventWriterOption { return []logEventWriterOption{} },
			errExpected: fmt.Errorf("Field validation for 'Logger' failed on the 'required' tag"),
		},
		"error on conifg": {
			v: v,
			opts: func(t *testing.T) []logEventWriterOption {
				return []logEventWriterOption{func(*logEventWriter) error { return fmt.Errorf("error on config") }}
			},
			errExpected: fmt.Errorf("error on config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q, err := NewLog(test.v, test.opts(t)...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, q)
		})
	}
}

func TestLogWriter_Write(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		loggerFn    func(*testing.T) (*zap.Logger, *observer.ObservedLogs)
		event       event.WaypointJobEvent
		errExpected error
	}{
		"happy path": {
			loggerFn: func(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
				core, logs := observer.New(zapcore.DebugLevel)
				return zap.New(core), logs
			},
			event: event.WaypointJobEvent{},
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger, logs := test.loggerFn(t)
			lw, err := NewLog(v, WithLog(logger))
			assert.NoError(t, err)

			err = lw.Write(test.event)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Len(t, logs.All(), 1)

			log := logs.All()[0]
			assert.Equal(t, zapcore.DebugLevel, log.Level)
			assert.Equal(t, "processed waypoint job event", log.Message)
			assert.True(t, log.Context[0].Equals(zap.Any("event", test.event)))
		})
	}
}
