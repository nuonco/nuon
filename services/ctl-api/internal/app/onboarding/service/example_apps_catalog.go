package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	exampleAppsCatalogURL   = "https://nuonco.github.io/example-app-configs/onboarding-apps.json"
	exampleAppsRepo         = "https://github.com/nuonco/example-app-configs"
	exampleAppsBranch       = "main"
	exampleAppsCacheTTL     = 5 * time.Minute
	exampleAppsFetchTimeout = 1 * time.Second
)

// ExampleApp represents a pre-configured example application available for onboarding.
type ExampleApp struct {
	Slug          string   `json:"slug"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	Tags          []string `json:"tags"`
	CloudProvider string   `json:"cloud_provider"`
	Directory     string   `json:"directory"`
}

// Catalog serves the example-apps list, lazily fetching from the public
// onboarding-apps.json on first request and caching for exampleAppsCacheTTL.
// Falls back to fallbackExampleApps if the remote fetch has never succeeded.
type Catalog struct {
	mu        sync.Mutex
	apps      []ExampleApp
	fetchedAt time.Time
	client    *http.Client
	l         *zap.Logger
}

type CatalogParams struct {
	fx.In

	L *zap.Logger
}

func NewCatalog(params CatalogParams) *Catalog {
	return &Catalog{
		client: &http.Client{Timeout: exampleAppsFetchTimeout},
		l:      params.L.With(zap.String("component", "example-apps-catalog")),
	}
}

// Get returns the catalog, refreshing from the remote source if the cache
// is stale or empty. Concurrent callers serialize on the internal mutex.
func (c *Catalog) Get(ctx context.Context) []ExampleApp {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.apps != nil && time.Since(c.fetchedAt) < exampleAppsCacheTTL {
		return c.apps
	}

	apps, err := c.fetch(ctx)
	if err != nil {
		c.l.Warn("example apps fetch failed", zap.Error(err))
		if c.apps != nil {
			return c.apps
		}
		return fallbackExampleApps
	}

	c.apps = apps
	c.fetchedAt = time.Now()
	return c.apps
}

func (c *Catalog) fetch(ctx context.Context) ([]ExampleApp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, exampleAppsCatalogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var apps []ExampleApp
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	if len(apps) == 0 {
		return nil, fmt.Errorf("catalog is empty")
	}
	return apps, nil
}

// fallbackExampleApps is a minimal compiled-in catalog used when the remote
// fetch has never succeeded. The authoritative catalog lives at
// https://nuonco.github.io/example-app-configs/onboarding-apps.json.
var fallbackExampleApps = []ExampleApp{
	{Slug: "httpbin", DisplayName: "HTTPBin", Description: "HTTPBin app on AWS EC2 using our minimal AWS Sandbox", Category: "architecture", Difficulty: "simple", Tags: []string{"ec2", "docker", "debugging"}, CloudProvider: "aws", Directory: "httpbin"},
	{Slug: "eks-simple", DisplayName: "EKS Simple", Description: "A simple Whoami HTTP service on AWS EKS", Category: "architecture", Difficulty: "simple", Tags: []string{"eks", "kubernetes", "alb", "certificate"}, CloudProvider: "aws", Directory: "eks-simple"},
}
