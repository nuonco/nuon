package credentials

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
)

// ContextKey is used to manage the credentials in the context
type ContextKey struct {
	ID string
}

// EnsureContext adds credentials into the context if they do not exist
func EnsureContext(ctx context.Context, cfg *Config) (context.Context, error) {
	if cfg.CacheID == "" {
		return nil, fmt.Errorf("no cache id set")
	}

	awsCfg, err := cfg.fetchCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch credentials: %w", err)
	}

	return context.WithValue(ctx, ContextKey{cfg.CacheID}, awsCfg), nil
}

// FromContext fetches credentials from the context
func FromContext(ctx context.Context, cfg *Config) (*aws.Config, error) {
	if cfg.CacheID == "" {
		return nil, fmt.Errorf("no cache id set")
	}

	val := ctx.Value(ContextKey{cfg.CacheID})
	if val == nil {
		return nil, fmt.Errorf("credentials not found: %s", cfg.CacheID)
	}

	creds, ok := val.(*aws.Config)
	if !ok {
		return nil, fmt.Errorf("invalid credentials found in context: %T", val)
	}

	return creds, nil
}
