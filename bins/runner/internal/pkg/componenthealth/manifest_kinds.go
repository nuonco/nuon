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

type ManifestKindsProvider struct {
	l       *zap.Logger
	cluster *ClusterProvider

	mu       sync.RWMutex
	gvks     map[string][]schema.GroupVersionKind
	loaded   bool
	objects  map[string]string
	releases map[string]string
}

type ManifestKindsProviderParams struct {
	fx.In

	L       *zap.Logger `name:"system"`
	Cluster *ClusterProvider
}

func NewManifestKindsProvider(params ManifestKindsProviderParams) *ManifestKindsProvider {
	return &ManifestKindsProvider{
		l:        params.L,
		cluster:  params.Cluster,
		gvks:     map[string][]schema.GroupVersionKind{},
		objects:  map[string]string{},
		releases: map[string]string{},
	}
}

func (p *ManifestKindsProvider) SetKinds(componentID string, gvks []schema.GroupVersionKind, objectKeys, releases []string) {
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
	for release, owner := range p.releases {
		if owner == componentID {
			delete(p.releases, release)
		}
	}
	for _, release := range releases {
		p.releases[release] = componentID
	}
	p.mu.Unlock()
	p.persist()
}

func (p *ManifestKindsProvider) ComponentForRelease(release string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	componentID, ok := p.releases[release]
	return componentID, ok
}

func (p *ManifestKindsProvider) ComponentForObject(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	componentID, ok := p.objects[key]
	return componentID, ok
}

func (p *ManifestKindsProvider) Load() {
	if p.cluster == nil {
		return
	}

	p.mu.Lock()
	p.loaded = true
	p.mu.Unlock()

	restored := map[string][]schema.GroupVersionKind{}
	restoredObjects := map[string]string{}
	restoredReleases := map[string]string{}
	for _, entry := range p.cluster.ComponentKinds() {
		if after, isObject := strings.CutPrefix(entry, objectEntryPrefix); isObject {
			componentID, key, found := strings.Cut(after, "|")
			if found && componentID != "" && key != "" {
				restoredObjects[key] = componentID
			}
			continue
		}
		if after, isRelease := strings.CutPrefix(entry, releaseEntryPrefix); isRelease {
			componentID, release, found := strings.Cut(after, "|")
			if found && componentID != "" && release != "" {
				restoredReleases[release] = componentID
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
	for release, componentID := range restoredReleases {
		if _, live := p.releases[release]; !live {
			p.releases[release] = componentID
		}
	}
}

const objectEntryPrefix = "obj:"

const releaseEntryPrefix = "rel:"

func (p *ManifestKindsProvider) persist() {
	if p.cluster == nil {
		return
	}

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
	for release, componentID := range p.releases {
		out = append(out, releaseEntryPrefix+componentID+"|"+release)
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
