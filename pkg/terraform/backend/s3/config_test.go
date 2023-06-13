package s3

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_s3_ConfigFile(t *testing.T) {
	v := validator.New()
	bucketCfg := generics.GetFakeObj[*BucketConfig]()

	staticCreds := generics.GetFakeObj[*credentials.Config]()
	staticCreds.UseDefault = false
	staticCreds.AssumeRole = nil

	assumeCreds := generics.GetFakeObj[*credentials.Config]()
	staticCreds.UseDefault = false
	assumeCreds.Static = nil

	defaultCreds := generics.GetFakeObj[*credentials.Config]()
	defaultCreds.UseDefault = true
	defaultCreds.Static = nil
	defaultCreds.AssumeRole = nil

	tests := map[string]struct {
		backendFn   func(*testing.T) *s3
		assertFn    func(*testing.T, []byte)
		errExpected error
	}{
		"creds backend": {
			backendFn: func(t *testing.T) *s3 {
				s, err := New(v,
					WithBucketConfig(bucketCfg),
					WithCredentials(staticCreds))
				assert.NoError(t, err)
				return s
			},
			assertFn: func(t *testing.T, byts []byte) {
				var resp map[string]interface{}
				err := json.Unmarshal(byts, &resp)
				assert.NoError(t, err)

				assert.Equal(t, staticCreds.Static.AccessKeyID, resp["access_key"])
				assert.Equal(t, staticCreds.Static.SecretAccessKey, resp["secret_key"])
				assert.Equal(t, staticCreds.Static.SessionToken, resp["token"])

				assert.Equal(t, resp["bucket"], bucketCfg.Name)
				assert.Equal(t, resp["key"], bucketCfg.Key)
				assert.Equal(t, resp["region"], bucketCfg.Region)

				_, ok := resp["role_arn"]
				assert.False(t, ok)
				_, ok = resp["session_timeout"]
				assert.False(t, ok)
				_, ok = resp["session_name"]
				assert.False(t, ok)
			},
			errExpected: nil,
		},
		"iam backend": {
			backendFn: func(t *testing.T) *s3 {
				s, err := New(v,
					WithBucketConfig(bucketCfg),
					WithCredentials(assumeCreds),
				)
				assert.NoError(t, err)
				return s
			},
			assertFn: func(t *testing.T, byts []byte) {
				var resp map[string]interface{}
				err := json.Unmarshal(byts, &resp)
				assert.NoError(t, err)

				assert.Equal(t, resp["role_arn"], assumeCreds.AssumeRole.RoleARN)
				assert.Equal(t, resp["session_name"], assumeCreds.AssumeRole.SessionName)

				assert.Equal(t, resp["bucket"], bucketCfg.Name)
				assert.Equal(t, resp["key"], bucketCfg.Key)
				assert.Equal(t, resp["region"], bucketCfg.Region)

				_, ok := resp["access_key"]
				assert.False(t, ok)
				_, ok = resp["secret_key"]
				assert.False(t, ok)
				_, ok = resp["token"]
				assert.False(t, ok)
			},
			errExpected: nil,
		},
		"default backend": {
			backendFn: func(t *testing.T) *s3 {
				s, err := New(v,
					WithBucketConfig(bucketCfg),
					WithCredentials(defaultCreds),
				)
				assert.NoError(t, err)
				return s
			},
			assertFn: func(t *testing.T, byts []byte) {
				var resp map[string]interface{}
				err := json.Unmarshal(byts, &resp)
				assert.NoError(t, err)

				assert.Equal(t, resp["bucket"], bucketCfg.Name)
				assert.Equal(t, resp["key"], bucketCfg.Key)
				assert.Equal(t, resp["region"], bucketCfg.Region)

				_, ok := resp["access_key"]
				assert.False(t, ok)
				_, ok = resp["secret_key"]
				assert.False(t, ok)
				_, ok = resp["token"]
				assert.False(t, ok)
				_, ok = resp["role_arn"]
				assert.False(t, ok)
				_, ok = resp["session_name"]
				assert.False(t, ok)
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			backend := test.backendFn(t)

			cfg, err := backend.ConfigFile(ctx)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, cfg)
		})
	}
}
