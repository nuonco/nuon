package componenthealth

import (
	"context"
	"fmt"
	"sync"
)

const (
	labelInstallID   = "nuon.co/install-id"
	labelComponentID = "nuon.co/component-id"

	helmManagedByLabel             = "app.kubernetes.io/managed-by"
	helmManagedByValue             = "Helm"
	helmReleaseNameAnnotation      = "meta.helm.sh/release-name"
	helmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
)

type componentEntry struct {
	installComponentID string
	componentID        string
	componentName      string
	componentType      string
	helmReleaseName    string
	helmNamespace      string
	probes             []probeSpec
}

type index struct {
	mu         sync.RWMutex
	installID  string
	components map[string]componentEntry
	helm       map[string]componentEntry
}

func newIndex() *index {
	return &index{components: map[string]componentEntry{}, helm: map[string]componentEntry{}}
}

func (i *index) replace(installID string, entries []componentEntry) {
	byComponentID := make(map[string]componentEntry, len(entries))
	byHelmRelease := map[string]componentEntry{}
	for _, e := range entries {
		if e.componentID == "" {
			continue
		}
		byComponentID[e.componentID] = e
		// why: Keyed by release name alone — the release namespace is often the install
		// namespace, not the component's config namespace, so it can't be in the key.
		if e.helmReleaseName != "" {
			byHelmRelease[e.helmReleaseName] = e
		}
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.installID = installID
	i.components = byComponentID
	i.helm = byHelmRelease
}

func (i *index) installIDValue() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.installID
}

func (i *index) lookup(componentID string) (componentEntry, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	e, ok := i.components[componentID]
	return e, ok
}

func (i *index) lookupHelm(releaseName string) (componentEntry, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	e, ok := i.helm[releaseName]
	return e, ok
}

func (i *index) componentsOfType(componentType string) []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	out := make([]string, 0, len(i.components))
	for id, e := range i.components {
		if e.componentType == componentType {
			out = append(out, id)
		}
	}
	return out
}

func (i *index) probeTargets() map[string][]probeSpec {
	i.mu.RLock()
	defer i.mu.RUnlock()

	out := map[string][]probeSpec{}
	for id, e := range i.components {
		if len(e.probes) > 0 {
			out[id] = e.probes
		}
	}
	return out
}

func (e *Engine) rebuildIndex(ctx context.Context) error {
	resp, err := e.apiClient.GetRunnerInstallComponents(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch install components: %w", err)
	}
	if resp == nil {
		return nil
	}

	entries := make([]componentEntry, 0, len(resp.Components))
	for _, c := range resp.Components {
		if c == nil {
			continue
		}
		entries = append(entries, componentEntry{
			installComponentID: c.InstallComponentID,
			componentID:        c.ComponentID,
			componentName:      c.ComponentName,
			componentType:      c.ComponentType,
			helmReleaseName:    c.HelmReleaseName,
			helmNamespace:      c.HelmNamespace,
			probes:             probeSpecsFor(c),
		})
	}

	e.idx.replace(resp.InstallID, entries)
	return nil
}
