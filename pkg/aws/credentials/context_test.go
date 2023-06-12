package credentials

import (
	"context"
	"fmt"
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	staticCreds := generics.GetFakeObj[StaticCredentials]()
	cacheID := generics.GetFakeObj[string]()

	tests := map[string]struct {
		cfgFn       func() *Config
		ctxFn       func(*testing.T, context.Context, *Config) context.Context
		assertFn    func(*testing.T, *aws.Credentials)
		errExpected error
	}{
		"happy path - creds in context": {
			cfgFn: func() *Config {
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
		"error - no creds in context": {
			cfgFn: func() *Config {
				return &Config{
					Static:  staticCreds,
					CacheID: cacheID,
				}
			},
			ctxFn: func(t *testing.T, ctx context.Context, cfg *Config) context.Context {
				return ctx
			},
			errExpected: fmt.Errorf("credentials not found"),
		},
		"error - invalid object in context": {
			cfgFn: func() *Config {
				return &Config{
					Static:  staticCreds,
					CacheID: cacheID,
				}
			},
			ctxFn: func(t *testing.T, ctx context.Context, cfg *Config) context.Context {
				ctx = context.WithValue(ctx, ContextKey{cfg.CacheID}, map[string]string{"key": "value"})
				return ctx
			},
			errExpected: fmt.Errorf("invalid credentials"),
		},
		"error - no cache id": {
			cfgFn: func() *Config {
				return &Config{
					Static: staticCreds,
				}
			},
			ctxFn: func(t *testing.T, ctx context.Context, cfg *Config) context.Context {
				return ctx
			},
			errExpected: fmt.Errorf("no cache id"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)
			defer cancelFn()

			cfg := test.cfgFn()
			ctx = test.ctxFn(t, ctx, cfg)

			awsCfg, err := FromContext(ctx, cfg)
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

func TestEnsureContext(t *testing.T) {
	staticCreds := generics.GetFakeObj[StaticCredentials]()
	cacheID := generics.GetFakeObj[string]()

	tests := map[string]struct {
		cfgFn       func() *Config
		assertFn    func(*testing.T, *Config, context.Context)
		errExpected error
	}{
		"happy path - creds are set": {
			cfgFn: func() *Config {
				return &Config{
					Static:  staticCreds,
					CacheID: cacheID,
				}
			},
			assertFn: func(t *testing.T, cfg *Config, ctx context.Context) {
				val := ctx.Value(ContextKey{cfg.CacheID})
				assert.NotNil(t, val)
				awsCfg, ok := val.(aws.Config)
				assert.True(t, ok)
				creds, err := awsCfg.Credentials.Retrieve(ctx)
				assert.NoError(t, err)
				assert.Equal(t, staticCreds.AccessKeyID, creds.AccessKeyID)
				assert.Equal(t, staticCreds.SecretAccessKey, creds.SecretAccessKey)
				assert.Equal(t, staticCreds.SessionToken, creds.SessionToken)
			},
		},
		"error - no cache ID set": {
			cfgFn: func() *Config {
				return &Config{
					Static: staticCreds,
				}
			},
			errExpected: fmt.Errorf("no cache id set"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancelFn := context.WithCancel(ctx)
			defer cancelFn()

			cfg := test.cfgFn()

			ctx, err := EnsureContext(ctx, cfg)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, cfg, ctx)
		})
	}
}
