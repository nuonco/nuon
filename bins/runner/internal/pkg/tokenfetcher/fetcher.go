package tokenfetcher

import (
	"context"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

type TokenFetchResult struct {
	RunnerID   string
	InstanceID string
	AccountID  string
	ProjectID  string
	Token      string
}

type TokenFetcher interface {
	FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*TokenFetchResult, error)

	Name() string
}
