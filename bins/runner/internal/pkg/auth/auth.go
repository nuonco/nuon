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

type Token struct {
	Value string
}

type Params struct {
	fx.In

	Cfg *runnerconfig.Config
	L   *zap.Logger `name:"dev"`
}

func New(params Params) (*Token, error) {
	if params.Cfg.RunnerAPIToken != "" {
		params.L.Info("using runner API token from config/env")
		return &Token{Value: params.Cfg.RunnerAPIToken}, nil
	}

	params.L.Info("fetching runner API token via IMDS",
		zap.String("platform", params.Cfg.RunnerPlatform))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(params.Cfg.RunnerAPIURL),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create unauthenticated client for token fetch: %w", err)
	}

	var result *fetchtoken.FetchTokenResult
	switch params.Cfg.RunnerPlatform {
	case "azure":
		result, err = fetchtoken.FetchTokenAzure(ctx, apiClient, params.Cfg.RunnerID)
	default:
		result, err = fetchtoken.FetchToken(ctx, apiClient, params.Cfg.RunnerAuthMethod, params.Cfg.RunnerID)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to fetch runner token via IMDS: %w", err)
	}

	params.Cfg.RunnerAPIToken = result.Token

	params.L.Info("successfully fetched runner API token",
		zap.String("runner_id", result.RunnerID),
		zap.String("instance_id", result.InstanceID),
	)

	return &Token{Value: result.Token}, nil
}
