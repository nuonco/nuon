package kube

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigForCluster(t *testing.T) {
	tests := map[string]struct {
		info        ClusterInfo
		errExpected error
	}{
		"happy path": {
			info: ClusterInfo{
				ID:             "cluster id",
				Endpoint:       "https://some.kubernetes.cluster",
				TrustedRoleARN: "arn:aws:something",
				CAData:         "YjY0IGVuY29kZWQgZGF0YQ==",
			},
		},

		"errors without ID": {
			info: ClusterInfo{
				Endpoint:       "https://some.kubernetes.cluster",
				TrustedRoleARN: "arn:aws:something",
				CAData:         "YjY0IGVuY29kZWQgZGF0YQ==",
			},
			errExpected: ErrInvalidCluster,
		},

		"errors without endpoint": {
			info: ClusterInfo{
				ID:             "cluster id",
				TrustedRoleARN: "arn:aws:something",
				CAData:         "YjY0IGVuY29kZWQgZGF0YQ==",
			},
			errExpected: ErrInvalidCluster,
		},

		"errors without cert data": {
			info: ClusterInfo{
				ID:             "cluster id",
				Endpoint:       "https://some.kubernetes.cluster",
				TrustedRoleARN: "arn:aws:something",
			},
			errExpected: ErrInvalidCert,
		},

		"errors when cert data isn't b64 encoded": {
			info: ClusterInfo{
				ID:             "cluster id",
				Endpoint:       "https://some.kubernetes.cluster",
				TrustedRoleARN: "arn:aws:something",
				CAData:         "not b64 encoded",
			},
			errExpected: ErrInvalidCert,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			origCA := test.info.CAData
			cfg, err := ConfigForCluster(&test.info)
			if test.errExpected != nil {
				assert.ErrorIs(t, err, test.errExpected)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			// data is decoded if there wasn't any error
			// so re-encode and make sure there's not issues
			d := base64.StdEncoding.EncodeToString([]byte(test.info.CAData))
			assert.Equal(t, origCA, d)
		})
	}
}
