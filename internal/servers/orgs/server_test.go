package orgs

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	orgsservice "github.com/powertoolsdev/orgs-api/internal/services/orgs"
	"gotest.tools/assert"
)

func TestNew(t *testing.T) {
	mockCtl := gomock.NewController(t)
	mockSvc := orgsservice.NewMockService(mockCtl)
	mockCtxProvider := orgcontext.NewMockProvider(mockCtl)

	tests := map[string]struct {
		optFns      func() []serverOption
		assertFn    func(*testing.T, *server)
		errExpected error
	}{
		"happy path": {
			optFns: func() []serverOption {
				return []serverOption{
					WithService(mockSvc),
					WithContextProvider(mockCtxProvider),
				}
			},
			assertFn: func(t *testing.T, srv *server) {
				assert.Equal(t, mockSvc, srv.Svc)
				assert.Equal(t, mockCtxProvider, srv.CtxProvider)
			},
		},
		"missing context provider": {
			optFns: func() []serverOption {
				return []serverOption{
					WithService(mockSvc),
				}
			},
			errExpected: fmt.Errorf("server.CtxProvider"),
		},
		"missing service": {
			optFns: func() []serverOption {
				return []serverOption{
					WithContextProvider(mockCtxProvider),
				}
			},
			errExpected: fmt.Errorf("server.Svc"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			opts := test.optFns()
			srv, err := New(opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, srv)
		})
	}
}
