package helm

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Resource is one object from a rendered Helm manifest.
// An empty Namespace means the manifest did not set one. Callers that know the
// kind is namespaced should use the release namespace instead.
type Resource struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

// ResourcesFromManifest reads objects out of a multi-document Helm manifest.
func ResourcesFromManifest(manifest string) ([]Resource, error) {
	if strings.TrimSpace(manifest) == "" {
		return nil, nil
	}

	dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
	var resources []Resource
	for {
		var obj unstructured.Unstructured
		if err := dec.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to parse rendered manifest: %w", err)
		}
		if len(obj.Object) == 0 || obj.GetKind() == "" || obj.GetName() == "" {
			continue
		}
		resources = append(resources, Resource{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Namespace:  obj.GetNamespace(),
			Name:       obj.GetName(),
		})
	}
	return resources, nil
}

// namespaceOr is the namespace a namespaced object will be installed into.
// Cluster-scoped objects should not use this.
func (r Resource) namespaceOr(releaseNamespace string) string {
	if r.Namespace != "" {
		return r.Namespace
	}
	if releaseNamespace != "" {
		return releaseNamespace
	}
	return "default"
}
