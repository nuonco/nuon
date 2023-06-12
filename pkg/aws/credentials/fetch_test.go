package credentials

import (
	"context"
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	staticCreds := generics.GetFakeObj[StaticCredentials]()
	cacheID := generics.GetFakeObj[string]()

	tests := map[string]struct {
		configFn    func() *Config
		ctxFn       func(*testing.T, context.Context, *Config) context.Context
		assertFn    func(*testing.T, *aws.Credentials)
		errExpected error
	}{
		"happy path - creds in context": {
			configFn: func() *Config {
				return &Config{
					Static:  staticCreds,
					CacheID: cacheID,
				}
			},
			ctxFn: func(t *testing.T, ctx context.Context, cfg *Config) context.Context {
				ctx, err := EnsureContext(ctx, cfg)
				assert.NoError(t, err)
				return ctx
			},
			assertFn: func(t *testing.T, creds *aws.Credentials) {
				assert.Equal(t, staticCreds.AccessKeyID, creds.AccessKeyID)
				assert.Equal(t, staticCreds.SecretAccessKey, creds.SecretAccessKey)
				assert.Equal(t, staticCreds.SessionToken, creds.SessionToken)
			},
		},
		"happy path - no creds in context": {
			configFn: func() *Config {
				return &Config{
					Static: staticCreds,
				}
			},
			ctxFn: func(t *testing.T, ctx context.Context, cfg *Config) context.Context {
				return ctx
			},
			assertFn: func(t *testing.T, creds *aws.Credentials) {
				assert.Equal(t, staticCreds.AccessKeyID, creds.AccessKeyID)
				assert.Equal(t, staticCreds.SecretAccessKey, creds.SecretAccessKey)
				assert.Equal(t, staticCreds.SessionToken, creds.SessionToken)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)
			defer cancelFn()

			cfg := test.configFn()
			ctx = test.ctxFn(t, ctx, cfg)

			awsCfg, err := Fetch(ctx, cfg)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			creds, err := awsCfg.Credentials.Retrieve(ctx)
			assert.NoError(t, err)
			test.assertFn(t, &creds)
		})
	}
}
