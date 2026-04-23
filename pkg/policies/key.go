// Package policies derives stable filenames for Kubernetes policy manifests
// shipped through the Nuon sandbox Terraform module.
//
// The sandbox module's `kubectl_manifest.vendor_policies` resource uses
// `for_each = fileset(...)`, so the filename is the Terraform state key.
// Anchoring the filename to the manifest's kind/namespace/name keeps the
// Terraform address bound to actual K8s object identity instead of the
// position of the policy in the source list.
package policies

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

var nonSlugChar = regexp.MustCompile(`[^a-z0-9]+`)

// LegacyKeyPattern matches the historic positional filenames (`0.yaml`,
// `12.yaml`) emitted before content-derived keys existed. The runner uses
// it to detect installs that still need a one-time `terraform state mv`
// migration.
var LegacyKeyPattern = regexp.MustCompile(`^\d+\.yaml$`)

// IsLegacyKey reports whether key matches the historic positional form.
func IsLegacyKey(key string) bool {
	return LegacyKeyPattern.MatchString(key)
}

// ManifestKeyFromYAML parses a single Kubernetes manifest YAML document and
// returns a deterministic filename-style key derived from the manifest's
// kind, namespace (when present), and metadata.name.
func ManifestKeyFromYAML(contents string) (string, error) {
	var manifest map[string]any
	if err := yaml.Unmarshal([]byte(contents), &manifest); err != nil {
		return "", fmt.Errorf("unable to parse manifest yaml: %w", err)
	}
	return ManifestKey(manifest)
}

// ManifestKey returns the deterministic key for an already-parsed manifest.
func ManifestKey(manifest map[string]any) (string, error) {
	kind, _ := manifest["kind"].(string)
	if kind == "" {
		return "", errors.New("manifest missing 'kind'")
	}

	md, ok := manifest["metadata"].(map[string]any)
	if !ok {
		// yaml.v2 decodes nested maps as `map[interface{}]interface{}`;
		// fall back to that shape before giving up.
		mdAny, _ := manifest["metadata"].(map[any]any)
		if mdAny == nil {
			return "", errors.New("manifest missing 'metadata'")
		}
		md = make(map[string]any, len(mdAny))
		for k, v := range mdAny {
			if ks, ok := k.(string); ok {
				md[ks] = v
			}
		}
	}
	name, _ := md["name"].(string)
	if name == "" {
		return "", errors.New("manifest missing 'metadata.name'")
	}

	parts := []string{kind, name}
	if ns, _ := md["namespace"].(string); ns != "" {
		parts = []string{kind, ns, name}
	}

	slug := strings.ToLower(strings.Join(parts, "-"))
	slug = nonSlugChar.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug + ".yaml", nil
}
