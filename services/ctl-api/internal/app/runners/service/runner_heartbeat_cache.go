package service

import (
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	runnerHeartbeatRunnerCacheSize  = 4096
	runnerHeartbeatInstallCacheSize = 4096
	runnerHeartbeatCacheTTL         = 1 * time.Hour
)

// RunnerHeartbeatCache holds the lookups used to enrich heartbeat metric tags.
// Renames (org, install) take up to the TTL to propagate.
type RunnerHeartbeatCache struct {
	Runners  *expirable.LRU[string, *app.Runner]
	Installs *expirable.LRU[string, *app.Install]
}

func NewRunnerHeartbeatCache() *RunnerHeartbeatCache {
	return &RunnerHeartbeatCache{
		Runners:  expirable.NewLRU[string, *app.Runner](runnerHeartbeatRunnerCacheSize, nil, runnerHeartbeatCacheTTL),
		Installs: expirable.NewLRU[string, *app.Install](runnerHeartbeatInstallCacheSize, nil, runnerHeartbeatCacheTTL),
	}
}
