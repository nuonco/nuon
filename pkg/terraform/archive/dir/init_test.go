package dir

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_dir_Init(t *testing.T) {
	errInit := fmt.Errorf("does not exist")

	tests := map[string]struct {
		dirFn       func(t *testing.T) string
		errExpected error
	}{
		"happy path": {
			dirFn: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
		},
		"error": {
			dirFn: func(t *testing.T) string {
				return "/tmp/does-not-exist"
			},
			errExpected: errInit,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			tmpDir := test.dirFn(t)
			obj := &dir{
				Path: tmpDir,
			}

			err := obj.Init(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
