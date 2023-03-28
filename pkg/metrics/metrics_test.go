package metrics

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNew(t *testing.T) {
	log := zaptest.NewLogger(t)

	tests := map[string]struct {
		optFns      func() []writerOption
		assertFn    func(*testing.T, *writer)
		errExpected error
	}{
		"defaults": {
			optFns: func() []writerOption {
				return []writerOption{
					WithLogger(log),
				}
			},
			assertFn: func(t *testing.T, w *writer) {
				assert.False(t, w.Disable)
				assert.Equal(t, defaultAddress, w.Address)
			},
		},
		"disable": {
			optFns: func() []writerOption {
				return []writerOption{
					WithLogger(log),
					WithDisable(true),
				}
			},
			assertFn: func(t *testing.T, w *writer) {
				assert.True(t, w.Disable)
			},
		},
		"tags": {
			optFns: func() []writerOption {
				return []writerOption{
					WithLogger(log),
					WithTags("key:value"),
				}
			},
			assertFn: func(t *testing.T, w *writer) {
				assert.Equal(t, w.Tags[0], "key:value")
			},
		},
		"invalid tag": {
			optFns: func() []writerOption {
				return []writerOption{
					WithLogger(log),
					WithTags("keyvalue"),
				}
			},
			errExpected: fmt.Errorf("invalid tag"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			r, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, r)
		})
	}
}
