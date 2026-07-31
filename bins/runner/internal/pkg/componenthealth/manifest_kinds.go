package componenthealth

import (
	"sort"
	"strings"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

// ManifestKindsProvider records the kinds each chart component rendered, handed over by
// the deploy job.
//
// Discovery cannot run the other way round: to find a release's objects the
// engine must already know which kinds to list, and the rendered manifest lives
// in the release Secret, which health's identity is deliberately denied. The
// deploy already holds the manifest, so it is the one place that knows.
//
// Fed by both helm and kubernetes_manifest deploys. Reading it at deploy time
// also makes it independent of helm's storage driver (secret, configmap, or
// nuon's own).
type ManifestKindsProvider struct {
	l       *zap.Logger
	cluster *ClusterProvider

	mu     sync.RWMutex
	gvks   map[string][]schema.GroupVersionKind
	loaded bool
	// objects maps a resourceKey to the component that applied it. Terraform
	// objects carry neither nuon labels nor helm annotations, so this is the only
	// record of who owns them — and it has to outlive the process.
	objects map[string]string
}

type ManifestKindsProviderParams struct {
	fx.In

	L       *zap.Logger `name:"system"`
	Cluster *ClusterProvider
}

func NewManifestKindsProvider(params ManifestKindsProviderParams) *ManifestKindsProvider {
	return &ManifestKindsProvider{
		l:       params.L,
		cluster: params.Cluster,
		gvks:    map[string][]schema.GroupVersionKind{},
		objects: map[string]string{},
	}
}

// SetKinds records kinds a caller already resolved (terraform state), plus the
// object keys that component owns.
func (p *ManifestKindsProvider) SetKinds(componentID string, gvks []schema.GroupVersionKind, objectKeys []string) {
	if componentID == "" {
		return
	}

	p.mu.Lock()
	if len(gvks) == 0 {
		delete(p.gvks, componentID)
	} else {
		p.gvks[componentID] = gvks
	}
	for key, owner := range p.objects {
		if owner == componentID {
			delete(p.objects, key)
		}
	}
	for _, key := range objectKeys {
		p.objects[key] = componentID
	}
	p.mu.Unlock()
	p.persist()
}

// ComponentForObject returns the component that applied an object, keyed as
// resourceKey(kind, namespace, name).
func (p *ManifestKindsProvider) ComponentForObject(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	componentID, ok := p.objects[key]
	return componentID, ok
}

// Load rehydrates kinds persisted by earlier deploys of this install.
func (p *ManifestKindsProvider) Load() {
	if p.cluster == nil {
		return
	}

	p.mu.Lock()
	p.loaded = true
	p.mu.Unlock()

	restored := map[string][]schema.GroupVersionKind{}
	restoredObjects := map[string]string{}
	for _, entry := range p.cluster.ComponentKinds() {
		if after, isObject := strings.CutPrefix(entry, objectEntryPrefix); isObject {
			componentID, key, found := strings.Cut(after, "|")
			if found && componentID != "" && key != "" {
				restoredObjects[key] = componentID
			}
			continue
		}
		componentID, gvk, ok := decodeComponentKind(entry)
		if !ok {
			continue
		}
		restored[componentID] = append(restored[componentID], gvk)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	for componentID, gvks := range restored {
		if _, live := p.gvks[componentID]; !live {
			p.gvks[componentID] = gvks
		}
	}
	for key, componentID := range restoredObjects {
		if _, live := p.objects[key]; !live {
			p.objects[key] = componentID
		}
	}
}

// objectEntryPrefix marks a persisted entry as object ownership rather than a
// kind, so both share one stored list.
const objectEntryPrefix = "obj:"

// persist mirrors every recorded kind so a restart does not lose them. Encoded
// flat as "componentID|group/version/Kind" to keep the stored shape a plain
// string list.
func (p *ManifestKindsProvider) persist() {
	if p.cluster == nil {
		return
	}

	// A deploy can land before the engine has rehydrated. Persisting the
	// in-memory map alone would then drop every other component's kinds, so
	// load first — the stored list is the union across components.
	p.mu.RLock()
	loaded := p.loaded
	p.mu.RUnlock()
	if !loaded {
		p.Load()
	}

	p.mu.RLock()
	out := make([]string, 0, 16)
	for componentID, gvks := range p.gvks {
		for _, gvk := range gvks {
			out = append(out, componentID+"|"+gvk.GroupVersion().String()+"/"+gvk.Kind)
		}
	}
	for key, componentID := range p.objects {
		out = append(out, objectEntryPrefix+componentID+"|"+key)
	}
	p.mu.RUnlock()

	sort.Strings(out)
	p.cluster.SetComponentKinds(out)
}

func decodeComponentKind(entry string) (string, schema.GroupVersionKind, bool) {
	componentID, rest, found := strings.Cut(entry, "|")
	if !found || componentID == "" {
		return "", schema.GroupVersionKind{}, false
	}
	idx := strings.LastIndex(rest, "/")
	if idx <= 0 || idx == len(rest)-1 {
		return "", schema.GroupVersionKind{}, false
	}
	gv, err := schema.ParseGroupVersion(rest[:idx])
	if err != nil {
		return "", schema.GroupVersionKind{}, false
	}
	return componentID, gv.WithKind(rest[idx+1:]), true
}

// Set replaces the kinds recorded for a component from a freshly rendered
// release manifest. An empty manifest clears them.
func (p *ManifestKindsProvider) Set(componentID, manifest string) {
	if componentID == "" {
		return
	}

	gvks := gvksFromManifest(manifest)

	p.mu.Lock()
	if len(gvks) == 0 {
		delete(p.gvks, componentID)
	} else {
		p.gvks[componentID] = gvks
	}
	p.mu.Unlock()
	p.persist()
}

// DiscoveredGVKs returns every kind the recorded releases rendered.
func (p *ManifestKindsProvider) DiscoveredGVKs() []schema.GroupVersionKind {
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
