package policies

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

var nonSlugChar = regexp.MustCompile(`[^a-z0-9]+`)

var LegacyKeyPattern = regexp.MustCompile(`^\d+\.yaml$`)

func IsLegacyKey(key string) bool {
	return LegacyKeyPattern.MatchString(key)
}

func ManifestKeyFromYAML(contents string) (string, error) {
	var manifest map[string]any
	if err := yaml.Unmarshal([]byte(contents), &manifest); err != nil {
		return "", fmt.Errorf("unable to parse manifest yaml: %w", err)
	}
	return ManifestKey(manifest)
}

func ManifestKey(manifest map[string]any) (string, error) {
	kind, _ := manifest["kind"].(string)
	if kind == "" {
		return "", errors.New("manifest missing 'kind'")
	}

	md, ok := manifest["metadata"].(map[string]any)
	if !ok {
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
