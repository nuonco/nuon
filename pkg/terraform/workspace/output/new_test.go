package output

import (
	"fmt"
	"io"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()
	lg := hclog.New(&hclog.LoggerOptions{
		Output: io.Discard,
	})

	tests := map[string]struct {
		errExpected error
		optsFn      func() []dualOption
		assertFn    func(*testing.T, *dual)
	}{
		"happy path": {
			optsFn: func() []dualOption {
				return []dualOption{
					WithLogger(lg),
				}
			},
			assertFn: func(t *testing.T, d *dual) {
				assert.Equal(t, lg, d.Logger)
			},
		},
		"missing log": {
			optsFn: func() []dualOption {
				return []dualOption{}
			},
			errExpected: fmt.Errorf("Logger"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			e, err := New(v, test.optsFn()...)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, e)
		})
	}
}
