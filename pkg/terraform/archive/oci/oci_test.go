package oci

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	auth := generics.GetFakeObj[*Auth]()
	img := generics.GetFakeObj[*Image]()
	v := validator.New()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []ociOption
		assertFn    func(*testing.T, *oci)
	}{
		"happy path": {
			optsFn: func() []ociOption {
				return []ociOption{
					WithAuth(auth),
					WithImage(img),
				}
			},
			assertFn: func(t *testing.T, s *oci) {
				assert.Equal(t, auth, s.Auth)
				assert.Equal(t, img, s.Image)
			},
		},
		"missing auth": {
			optsFn: func() []ociOption {
				return []ociOption{
					WithImage(img),
				}
			},
			errExpected: fmt.Errorf("Auth"),
		},
		"missing image": {
			optsFn: func() []ociOption {
				return []ociOption{
					WithAuth(auth),
				}
			},
			errExpected: fmt.Errorf("Image"),
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
