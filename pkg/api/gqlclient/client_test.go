package gqlclient

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	url := generics.GetFakeObj[string]()
	token := generics.GetFakeObj[string]()

	tests := map[string]struct {
		errExpected error
		optsFn      func() []clientOption
		assertFn    func(*testing.T, *client)
	}{
		"happy path": {
			optsFn: func() []clientOption {
				return []clientOption{
					WithAuthToken(token),
					WithURL(url),
				}
			},
			assertFn: func(t *testing.T, c *client) {
				assert.Equal(t, url, c.URL)
				assert.Equal(t, token, c.AuthToken)
			},
		},
		"missing url": {
			optsFn: func() []clientOption {
				return []clientOption{
					WithAuthToken(token),
				}
			},
			errExpected: fmt.Errorf("URL"),
		},
		"missing auth token": {
			optsFn: func() []clientOption {
				return []clientOption{
					WithURL(url),
				}
			},
			errExpected: fmt.Errorf("AuthToken"),
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
