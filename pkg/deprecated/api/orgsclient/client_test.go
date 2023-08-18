package orgsclient

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := validator.New()

	addr := generics.GetFakeObj[string]()
	httpClient := http.DefaultClient

	tests := map[string]struct {
		errExpected error
		optsFn      func() []apiOption
		assertFn    func(*testing.T, *Client)
	}{
		"happy path": {
			optsFn: func() []apiOption {
				return []apiOption{
					WithAddr(addr),
				}
			},
			assertFn: func(t *testing.T, e *Client) {
				assert.Equal(t, addr, e.Addr)
				assert.NotNil(t, e.HttpClient)
				assert.NotNil(t, e.Apps)
				assert.NotNil(t, e.Builds)
				assert.NotNil(t, e.Installs)
				assert.NotNil(t, e.Instances)
				assert.NotNil(t, e.Orgs)
			},
		},
		"missing address": {
			optsFn: func() []apiOption {
				return []apiOption{}
			},
			errExpected: fmt.Errorf("Addr"),
		},
		"happy path - sets http client": {
			optsFn: func() []apiOption {
				return []apiOption{
					WithAddr(addr),
					WithHTTPClient(httpClient),
				}
			},
			assertFn: func(t *testing.T, e *Client) {
				assert.Equal(t, httpClient, e.HttpClient)
			},
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
