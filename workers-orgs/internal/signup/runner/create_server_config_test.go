package runner

import (
	"fmt"
	"testing"

	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func TestCreateServerConfig_validateRequest(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() CreateServerConfigRequest
		errExpected error
	}{
		"happy path": {
			reqFn: generics.GetFakeObj[CreateServerConfigRequest],
		},
		"no-org-id": {
			reqFn: func() CreateServerConfigRequest {
				req := generics.GetFakeObj[CreateServerConfigRequest]()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("CreateServerConfigRequest.OrgID"),
		},
		"no-namespace": {
			reqFn: func() CreateServerConfigRequest {
				req := generics.GetFakeObj[CreateServerConfigRequest]()
				req.TokenSecretNamespace = ""
				return req
			},
			errExpected: fmt.Errorf("CreateServerConfigRequest.TokenSecretNamespace"),
		},
		"no-server-addr": {
			reqFn: func() CreateServerConfigRequest {
				req := generics.GetFakeObj[CreateServerConfigRequest]()
				req.OrgServerAddr = ""
				return req
			},
			errExpected: fmt.Errorf("CreateServerConfigRequest.OrgServerAddr"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			err := req.validate()
			if test.errExpected == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, test.errExpected.Error())
			}
		})
	}
}
