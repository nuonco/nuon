package kubernetes_manifest

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "sigs.k8s.io/yaml"
)

func ResourcesToMultiDocYAML(objects []*unstructured.Unstructured) (string, error) {
	if len(objects) == 0 {
		return "", nil
	}

	var docs []string
	for _, obj := range objects {
		if obj == nil {
			continue
		}

		yamlBytes, err := k8syaml.Marshal(obj.Object)
		if err != nil {
			return "", err
		}

		docs = append(docs, string(yamlBytes))
	}

	return strings.Join(docs, "---\n"), nil
}

func kubernetesResourcesToMultiDocYAML(resources []*kubernetesResource) (string, error) {
	objects := make([]*unstructured.Unstructured, 0, len(resources))
	for _, r := range resources {
		if r != nil {
			objects = append(objects, r.obj)
		}
	}
	return ResourcesToMultiDocYAML(objects)
}
