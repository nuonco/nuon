package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	cliLatestToken    = "latest"
	cliLatestURL      = "https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/latest.txt"
	cliLatestCacheTTL = time.Minute
	cliLatestTimeout  = 3 * time.Second
)

type latestCLICache struct {
	mu        sync.Mutex
	version   string
	fetchedAt time.Time
	client    *http.Client
}

func (s *service) resolveMinCLIVersion(ctx context.Context) string {
	if s.cfg == nil {
		return ""
	}
	configured := strings.TrimSpace(s.cfg.MinCLIVersion)
	if !strings.EqualFold(configured, cliLatestToken) {
		return s.cfg.MinCLIVersion
	}

	version, err := s.latestCLI.resolve(ctx)
	if err != nil && s.l != nil {
		s.l.Warn("fetching latest cli version", zap.Error(err), zap.String("served", version))
	}
	return version
}

func (c *latestCLICache) resolve(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if !c.fetchedAt.IsZero() && now.Sub(c.fetchedAt) < cliLatestCacheTTL {
		return c.version, nil
	}

	if c.client == nil {
		c.client = &http.Client{Timeout: cliLatestTimeout}
	}

	version, err := fetchCLILatest(ctx, c.client)
	if err != nil {
		if ctx.Err() == nil {
			c.fetchedAt = now
		}
		return c.version, err
	}

	c.version = version
	c.fetchedAt = now
	return c.version, nil
}

func fetchCLILatest(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cliLatestURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("latest cli version returned %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(body))
	if version == "" {
		return "", fmt.Errorf("latest cli version is empty")
	}
	return version, nil
}
