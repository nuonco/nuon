package auth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal"
	fetchtoken "github.com/nuonco/nuon/bins/runner/internal/jobs/management/fetch_token"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// Token holds the runner API token fetched during initialization.
// It is provided as an fx dependency so that downstream providers (e.g. api.New)
// can depend on it and be guaranteed the token is available.
type Token struct {
	Value string
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger `name:"dev"`
}

func New(params Params) (*Token, error) {
	// If the token is already set (e.g. local dev via env var), use it directly.
	if params.Cfg.RunnerAPIToken != "" {
		params.L.Info("using runner API token from config/env")
		return &Token{Value: params.Cfg.RunnerAPIToken}, nil
	}

	// TODO(fd): we need to branch here based on platform. but the platform comes from the api.
	// for now, we'll leave this as is. in the future, we'll likely autodetect the platform.
	params.L.Info("fetching runner API token via IMDS")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(params.Cfg.RunnerAPIURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create unauthenticated client for token fetch: %w", err)
	}

	result, err := fetchtoken.FetchToken(ctx, apiClient)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch runner token via IMDS: %w", err)
	}

	// Backfill config so existing code that reads cfg.RunnerAPIToken continues to work.
	params.Cfg.RunnerAPIToken = result.Token

	params.L.Info("successfully fetched runner API token",
		zap.String("runner_id", result.RunnerID),
		zap.String("instance_id", result.InstanceID),
	)

	return &Token{Value: result.Token}, nil
}
