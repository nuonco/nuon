package config

import (
	"sort"
	"strings"
)

// NormalizeIntermediateConfig makes intermediate AppConfig JSON deterministic so
// identical semantic configs do not produce false-positive diffs across runs.
func NormalizeIntermediateConfig(cfg *AppConfig) {
	if cfg == nil {
		return
	}

	if len(cfg.Components) > 1 {
		sort.SliceStable(cfg.Components, func(i, j int) bool {
			return cfg.Components[i].Name < cfg.Components[j].Name
		})
	}

	for _, comp := range cfg.Components {
		if comp == nil || comp.KubernetesManifest == nil {
			continue
		}
		km := comp.KubernetesManifest
		km.Manifest = normalizeManifestContent(km.Manifest)
		if km.Kustomize != nil && len(km.Kustomize.Patches) > 1 {
			sort.Strings(km.Kustomize.Patches)
		}
	}
}

func normalizeManifestContent(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimSpace(s)
}
