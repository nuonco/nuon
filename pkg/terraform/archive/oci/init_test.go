package oci

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_oci_Init(t *testing.T) {
	v := validator.New()
	img := generics.GetFakeObj[*Image]()
	auth := generics.GetFakeObj[*Auth]()

	tests := map[string]struct {
		tmpDirFn    func(*testing.T) string
		errExpected error
		assertFn    func(*testing.T, *oci)
	}{
		"happy path": {
			tmpDirFn: func(t *testing.T) string {
				return t.TempDir()
			},
			assertFn: func(t *testing.T, o *oci) {
				assert.NotNil(t, o.store)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)
			defer cancelFn()

			o, err := New(v, WithAuth(auth), WithImage(img))
			assert.NoError(t, err)
			o.tmpDir = test.tmpDirFn(t)

			err = o.Init(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, o)
		})
	}
}
