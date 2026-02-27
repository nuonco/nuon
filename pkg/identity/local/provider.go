package local

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/identity"
)

// FetcherConfig contains configuration for the local identity fetcher
type FetcherConfig struct {
	// RunnerID to use (optional, reads from env if not set)
	RunnerID string

	// Token to use (optional, reads from env if not set)
	Token string
}

// Fetcher is a stub fetcher for local development and testing
type Fetcher struct {
	config *FetcherConfig
}

// NewFetcher creates a new local identity fetcher
func NewFetcher(config *FetcherConfig) *Fetcher {
	return &Fetcher{
		config: config,
	}
}

// FetchToken returns a stub token for local development
func (f *Fetcher) FetchToken(ctx context.Context) (*identity.Token, error) {
	runnerID := f.config.RunnerID
	if runnerID == "" {
		runnerID = os.Getenv("RUNNER_ID")
	}
	if runnerID == "" {
		return nil, fmt.Errorf("no runner ID configured (set RUNNER_ID env var)")
	}

	token := f.config.Token
	if token == "" {
		token = os.Getenv("RUNNER_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("no token configured (set RUNNER_TOKEN env var)")
	}

	return &identity.Token{
		Raw:      token,
		Provider: identity.ProviderTypeLocal,
		RunnerID: runnerID,
		Metadata: map[string]string{
			"mode": "local",
		},
	}, nil
}

// Provider returns the provider type
func (f *Fetcher) Provider() identity.ProviderType {
	return identity.ProviderTypeLocal
}

// ValidatorConfig contains configuration for the local identity validator
type ValidatorConfig struct {
	// Logger for debug output
	Logger *zap.Logger

	// AllowAny allows any token (for testing only)
	AllowAny bool
}

// Validator is a stub validator for local development and testing
type Validator struct {
	logger   *zap.Logger
	allowAny bool
}

// NewValidator creates a new local identity validator
func NewValidator(config *ValidatorConfig) *Validator {
	logger := config.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Validator{
		logger:   logger,
		allowAny: config.AllowAny,
	}
}

// ValidateToken validates a local token (stub implementation)
func (v *Validator) ValidateToken(ctx context.Context, token string) (*identity.Claims, error) {
	if !v.allowAny && token == "" {
		return nil, fmt.Errorf("empty token not allowed")
	}

	v.logger.Debug("local token validation (stub)",
		zap.Bool("allow_any", v.allowAny),
		zap.Int("token_length", len(token)))

	return &identity.Claims{
		Subject:  "local-runner",
		Issuer:   "local",
		Audience: []string{"local"},
		Provider: identity.ProviderTypeLocal,
		Custom: map[string]interface{}{
			"mode": "local",
		},
	}, nil
}

// ValidateIdentity validates identity (stub - always passes)
func (v *Validator) ValidateIdentity(ctx context.Context, claims *identity.Claims, valCtx *identity.ValidationContext) error {
	v.logger.Debug("local identity validation (stub)",
		zap.String("runner_id", valCtx.RunnerID))
	return nil
}

// Provider returns the provider type
func (v *Validator) Provider() identity.ProviderType {
	return identity.ProviderTypeLocal
}

// IsAvailable always returns true for local mode
func IsAvailable(ctx context.Context) bool {
	return true
}
