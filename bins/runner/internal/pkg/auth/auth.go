package auth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// Token holds the runner API token fetched during initialization.
// It is provided as an fx dependency so that downstream providers (e.g. api.New)
// can depend on it and be guaranteed the token is available.
type Token struct {
	Value string
	// ExpiresAt is zero when the token is not refreshed in-process, e.g. one
	// supplied through config or issued without an expiry.
	ExpiresAt time.Time
}

type Params struct {
	fx.In

	Cfg *runnerconfig.Config
	L   *zap.Logger `name:"dev"`
}

func New(params Params) (*Token, error) {
	// If the token is already set (e.g. local dev via env var), use it directly.
	if token := params.Cfg.APIToken(); token != "" {
		params.L.Info("using runner API token from config/env")
		return &Token{Value: token}, nil
	}

	params.L.Info("fetching runner API token via IMDS",
		zap.String("platform", params.Cfg.RunnerPlatform))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := fetch(ctx, params.Cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch runner token via IMDS: %w", err)
	}

	// Backfill config so existing code that reads cfg.APIToken() continues to work.
	params.Cfg.SetAPIToken(result.Token)

	params.L.Info("successfully fetched runner API token",
		zap.String("runner_id", result.RunnerID),
		zap.String("instance_id", result.InstanceID),
		zap.Time("expires_at", result.ExpiresAt),
		zap.String("audience", result.Audience),
	)

	return &Token{Value: result.Token, ExpiresAt: result.ExpiresAt}, nil
}

func fetch(ctx context.Context, cfg *runnerconfig.Config) (*fetchtoken.FetchTokenResult, error) {
	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(cfg.RunnerAPIURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create unauthenticated client for token fetch: %w", err)
	}

	switch cfg.RunnerPlatform {
	case "azure":
		return fetchtoken.FetchTokenAzure(ctx, apiClient, cfg.RunnerID, true)
	default:
		return fetchtoken.FetchToken(ctx, apiClient, cfg.RunnerAuthMethod, cfg.RunnerID, true)
	}
}
