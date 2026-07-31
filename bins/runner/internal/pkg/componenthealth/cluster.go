package componenthealth

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/kube"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const persistTimeout = 15 * time.Second

// ClusterProvider holds the install's cluster access + the sandbox's managed
// helm releases, mirrored to ctl-api so the stateless engine rehydrates on boot
// (surviving restarts/image-swaps). ClusterInfo carries a durable assume-role
// config, so credentials are re-derived fresh on every use.
type ClusterProvider struct {
	apiClient nuonrunner.Client
	l         *zap.Logger

	mu              sync.RWMutex
	clusterInfo     *kube.ClusterInfo
	sandboxReleases map[string]struct{}
	// componentKinds is owned by ManifestKindsProvider; mirrored here because
	// this provider owns the single round trip to ctl-api.
	componentKinds []string
}

type ProviderParams struct {
	fx.In

	APIClient nuonrunner.Client
	L         *zap.Logger `name:"system"`
}

func NewClusterProvider(params ProviderParams) *ClusterProvider {
	return &ClusterProvider{apiClient: params.APIClient, l: params.L, sandboxReleases: map[string]struct{}{}}
}

func (p *ClusterProvider) Set(ci *kube.ClusterInfo) {
	if ci == nil {
		return
	}
	p.mu.Lock()
	p.clusterInfo = ci
	p.mu.Unlock()
	p.persist()
}

func (p *ClusterProvider) SetSandboxReleases(names []string) {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		if n != "" {
			set[n] = struct{}{}
		}
	}
	p.mu.Lock()
	p.sandboxReleases = set
	p.mu.Unlock()
	p.persist()
}

func (p *ClusterProvider) Get() *kube.ClusterInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.clusterInfo
}

func (p *ClusterProvider) IsSandboxRelease(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.sandboxReleases[name]
	return ok
}

// SetComponentKinds records the kinds discovered by deploys and persists them,
// so a restarted engine keeps watching them instead of narrowing back to the
// core workload kinds until every component is redeployed.
func (p *ClusterProvider) SetComponentKinds(kinds []string) {
	p.mu.Lock()
	p.componentKinds = append([]string(nil), kinds...)
	p.mu.Unlock()
	p.persist()
}

// ComponentKinds returns the persisted kinds, for rehydration on boot.
func (p *ClusterProvider) ComponentKinds() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]string(nil), p.componentKinds...)
}

// Load rehydrates the context from ctl-api on engine boot.
func (p *ClusterProvider) Load(ctx context.Context) {
	if p.apiClient == nil {
		return
	}

	clusterInfoJSON, releases, kinds, err := p.apiClient.GetComponentHealthContext(ctx)
	if err != nil {
		p.l.Warn("unable to load persisted component health context", zap.Error(err))
		return
	}

	var ci *kube.ClusterInfo
	if clusterInfoJSON != "" {
		ci = &kube.ClusterInfo{}
		if err := json.Unmarshal([]byte(clusterInfoJSON), ci); err != nil {
			p.l.Warn("unable to parse persisted cluster info", zap.Error(err))
			ci = nil
		}
	}
	set := make(map[string]struct{}, len(releases))
	for _, n := range releases {
		if n != "" {
			set[n] = struct{}{}
		}
	}

	p.mu.Lock()
	if ci != nil {
		p.clusterInfo = ci
	}
	if len(set) > 0 {
		p.sandboxReleases = set
	}
	if len(kinds) > 0 {
		p.componentKinds = kinds
	}
	p.mu.Unlock()
}

// persist mirrors the full current context to ctl-api, best-effort and off the
// caller's goroutine so it never slows or breaks a deploy.
func (p *ClusterProvider) persist() {
	// A background goroutine must never panic the runner on a missing client.
	if p.apiClient == nil {
		return
	}

	p.mu.RLock()
	ci := p.clusterInfo
	releases := make([]string, 0, len(p.sandboxReleases))
	for name := range p.sandboxReleases {
		releases = append(releases, name)
	}
	kinds := append([]string(nil), p.componentKinds...)
	p.mu.RUnlock()

	var clusterInfoJSON string
	if ci != nil {
		b, err := json.Marshal(ci)
		if err != nil {
			p.l.Warn("unable to marshal cluster info for persistence", zap.Error(err))
			return
		}
		clusterInfoJSON = string(b)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), persistTimeout)
		defer cancel()
		if err := p.apiClient.PutComponentHealthContext(ctx, clusterInfoJSON, releases, kinds); err != nil {
			p.l.Warn("unable to persist component health context", zap.Error(err))
		}
	}()
}
