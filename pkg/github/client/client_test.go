package client

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
)

// NOTE(jm): fake gh key generated using `openssl genrsa 1024`
const fakeGHKey string = `
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQC6XXN7LZest86JsvMx+z9oBfMOmiDtbPx2inZYQ+znH8K9Xkws
8k3KMjg4amSnN1/57VG40Ld9PQy9MPu0hwY90+4+ESdPbTFm7ge57lDXCzN85613
e9e7bPXZsvtqr2AdHuojjE573FXSIlrFs5YdfbgzsZEZH2lJN+s0+sSnaQIDAQAB
AoGAF7L2kn1zwkUFgME+5+Y5Y/MNu5eiBE9Ns41cC1Fn+OQzEX3CVhziA4prV9E/
x3vlOpURRV1VWBnBWvW1rHlXM6H4akZDPLlX5+WLEx5DNWfCV4r4HBw+Eusqal2C
OurqXREB4SfL/z6DVSofrKP6MnO1R/WzUCXdZlCB8451gGECQQDk9WvusXRIIqoV
WdBu7iyudyOQeybolWP/JqzlbA1Jk6tKKPGy2vE+elnjUUK4fwYHK417uOd5au3m
+ZdyrnxlAkEA0GA1ZJRb5RWUj9lTjg7wjYXgpE7pC9m+S20T3NbM/XFGieXOFrpn
1CXwXlfhHC2QTw4KsB9m4F7tWajQFH6ktQJAVzsuAQX5AJa8aGAMqobx3RKlKSHS
hCCOtaJ9kvck5NhvFVUeKP+DlEM3RgUqv3Id0NOGFxIulrLnqu2DBv11hQJAfxuA
1lata6NrWQgfsNMqJ5oXuwKro9/x9X6XFCovJxZ3Cd0VhsW0WjO+WT5QAelFUwPk
vySYk5s0O3H/Y9EQ1QJBAKFFeNyXeR5r2l5946Ore3jOKNs1HSMFEs80Gvg2cHVW
Q+q60PWlofEK84Z0WbLV2JEydT54sFBd20Ms0eWJts4=
-----END RSA PRIVATE KEY-----
`

func TestNew(t *testing.T) {
	appID := "987654321"
	httpClient := &http.Client{}

	tests := map[string]struct {
		optFns      func() []Option
		assertFn    func(*testing.T, *github.Client)
		errExpected error
	}{
		"happy path": {
			optFns: func() []Option {
				return []Option{
					WithAppID(appID),
					WithAppKey([]byte(fakeGHKey)),
				}
			},
			assertFn: func(*testing.T, *github.Client) {
			},
		},
		"missing app ID": {
			optFns: func() []Option {
				return []Option{
					WithAppKey([]byte(fakeGHKey)),
				}
			},
			errExpected: fmt.Errorf("AppID"),
		},
		"missing app key": {
			optFns: func() []Option {
				return []Option{
					WithAppID(appID),
				}
			},
			errExpected: fmt.Errorf("AppKey"),
		},
		"custom http client": {
			optFns: func() []Option {
				return []Option{
					WithAppKey([]byte(fakeGHKey)),
					WithAppID(appID),
					WithHTTPClient(httpClient),
				}
			},
			assertFn: func(t *testing.T, gh *github.Client) {
				assert.Equal(t, httpClient, gh.Client())
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			client, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}
