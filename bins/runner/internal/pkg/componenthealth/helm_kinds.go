package componenthealth

import (
	"strings"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// HelmProvider records the kinds each chart component rendered, handed over by
// the deploy job.
//
// Discovery cannot run the other way round: to find a release's objects the
// engine must already know which kinds to list, and the rendered manifest lives
// in the release Secret, which health's identity is deliberately denied. The
// deploy already holds the manifest, so it is the one place that knows.
type HelmProvider struct {
	l *zap.Logger

	mu   sync.RWMutex
	gvks map[string][]schema.GroupVersionKind
}

type HelmProviderParams struct {
	fx.In

	L *zap.Logger `name:"system"`
}

func NewHelmProvider(params HelmProviderParams) *HelmProvider {
	return &HelmProvider{l: params.L, gvks: map[string][]schema.GroupVersionKind{}}
}

// Set replaces the kinds recorded for a component from a freshly rendered
// release manifest. An empty manifest clears them.
func (p *HelmProvider) Set(componentID, manifest string) {
	if componentID == "" {
		return
	}

	gvks := gvksFromManifest(manifest)

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(gvks) == 0 {
		delete(p.gvks, componentID)
		return
	}
	p.gvks[componentID] = gvks
}

// DiscoveredGVKs returns every kind the recorded releases rendered.
func (p *HelmProvider) DiscoveredGVKs() []schema.GroupVersionKind {
	p.mu.RLock()
	defer p.mu.RUnlock()

	seen := map[schema.GroupVersionKind]struct{}{}
	out := make([]schema.GroupVersionKind, 0, 8)
	for _, gvks := range p.gvks {
		for _, gvk := range gvks {
			if _, dup := seen[gvk]; dup {
				continue
			}
			seen[gvk] = struct{}{}
			out = append(out, gvk)
		}
	}
	return out
}

// gvksFromManifest reads apiVersion/kind out of each document. Unparseable or
// incomplete documents are skipped: a missed kind costs coverage, a wrong one
// costs a failing list call every cycle.
func gvksFromManifest(manifest string) []schema.GroupVersionKind {
	var out []schema.GroupVersionKind
	seen := map[schema.GroupVersionKind]struct{}{}

	for _, doc := range strings.Split(manifest, "\n---") {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		var head struct {
			APIVersion string `json:"apiVersion"`
			Kind       string `json:"kind"`
		}
		if err := yaml.Unmarshal([]byte(doc), &head); err != nil {
			continue
		}
		if head.Kind == "" || head.APIVersion == "" {
			continue
		}
		gv, err := schema.ParseGroupVersion(head.APIVersion)
		if err != nil {
			continue
		}
		gvk := gv.WithKind(head.Kind)
		if _, dup := seen[gvk]; dup {
			continue
		}
		seen[gvk] = struct{}{}
		out = append(out, gvk)
	}
	return out
}
