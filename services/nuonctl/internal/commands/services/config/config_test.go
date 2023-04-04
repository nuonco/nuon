package config

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	rootDir := "/tmp/mono"
	v := validator.New()
	svcName := uuid.NewString()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []loaderOption
		assertFn    func(*testing.T, *loader)
	}{
		"happy path": {
			optsFn: func() []loaderOption {
				return []loaderOption{
					WithRootDir(rootDir),
					WithService(svcName),
				}
			},
			assertFn: func(t *testing.T, l *loader) {
				assert.Equal(t, rootDir, l.RootDir)
				assert.Equal(t, svcName, l.Service)
			},
		},
		"missing root dir": {
			optsFn: func() []loaderOption {
				return []loaderOption{
					WithService(svcName),
				}
			},
			errExpected: fmt.Errorf("RootDir"),
		},
		"missing service name": {
			optsFn: func() []loaderOption {
				return []loaderOption{
					WithRootDir(rootDir),
				}
			},
			errExpected: fmt.Errorf("Service"),
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
