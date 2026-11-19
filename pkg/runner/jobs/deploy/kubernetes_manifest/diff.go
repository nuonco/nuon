package kubernetes_manifest

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	// "k8s.io/utils/diff"
	"sigs.k8s.io/yaml"
)

const (
	// NOTE(jm): this can also be template or simple
	defaultOutputFormat string = "diff"
)

func resourceName(r *kubernetesResource) string {
	return fmt.Sprintf("%s-%s-%s-%s", r.groupVersionResource, r.groupVersionKind, r.namespace, r.name)
}

func resourcesToMap(in []*kubernetesResource) map[string]*kubernetesResource {
	o := make(map[string]*kubernetesResource, len(in))
	for _, r := range in {
		o[resourceName(r)] = r
	}
	return o
}

func (h *handler) resourceDiff(prev, curr []*kubernetesResource) ([]kubernetesResource, []kubernetesResource) {
	var delete []kubernetesResource
	var apply []kubernetesResource
	prevMap := resourcesToMap(prev)
	currMap := resourcesToMap(curr)

	for pid, p := range prevMap {
		if _, ok := currMap[pid]; !ok {
			delete = append(delete, *p)
		}
	}

	for _, c := range curr {
		apply = append(apply, *c)
	}

	return apply, delete
}

// objDiff compares two unstructured objects and returns a diff string.
// not being utilized at the moment, but can be used for debugging purposes
func (h *handler) objDiff(prev, curr unstructured.Unstructured) (string, error) {
	prev = h.removeManagedFields(&prev)
	curr = h.removeManagedFields(&curr)

	yamlA, err := yaml.Marshal(prev.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object A: %w", err)
	}

	yamlB, err := yaml.Marshal(curr.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object B: %w", err)
	}

	return cmp.Diff(string(yamlA), string(yamlB)), nil
}
