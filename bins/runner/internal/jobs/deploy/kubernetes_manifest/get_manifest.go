package kubernetes_manifest

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

func (h *handler) removeManagedFields(obj *unstructured.Unstructured) unstructured.Unstructured {
	clean := obj.DeepCopy()
	unstructured.RemoveNestedField(clean.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(clean.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(clean.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(clean.Object, "metadata", "generation")
	unstructured.RemoveNestedField(clean.Object, "metadata", "uid")
	unstructured.RemoveNestedField(clean.Object, "status")
	return *clean
}

func (h *handler) getKubernetesResourcesFromManifest(k *kubernetesClient, manifest string) ([]*kubernetesResource, error) {
	manifestRaw := strings.NewReader(manifest)
	dec := yaml.NewYAMLOrJSONDecoder(manifestRaw, 1024)
	var currentKubernetesResources []*kubernetesResource
	for {
		// parse the YAML doc
		o := map[string]interface{}{}
		err := dec.Decode(&o)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to decode kubernetes manifest %w", err)
		}

		raw, err := k8syaml.Marshal(o)
		if err != nil {
			return nil, fmt.Errorf("unable to encode resource to raw string %w", err)
		}

		obj := &unstructured.Unstructured{Object: o}

		gvk := obj.GroupVersionKind()
		restMapping, err := k.discoveryMapper.RESTMapping(
			gvk.GroupKind(),
			gvk.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get restmapping %w", err)
		}

		resourceList, err := k.resourcesforGroupVersion(fmt.Sprintf("%s/%s", restMapping.Resource.Group, gvk.Version))
		if err != nil {
			return nil, fmt.Errorf("unable to get resource list: %w", err)
		}
		namespaced := false
		for _, apiResource := range resourceList.APIResources {
			if apiResource.Kind == obj.GetKind() {
				namespaced = apiResource.Namespaced
			}
		}

		namespace := obj.GetNamespace()
		if len(namespace) == 0 {
			namespace = h.state.plan.KubernetesManifestDeployPlan.Namespace
		}

		currentKubernetesResources = append(currentKubernetesResources, &kubernetesResource{
			groupVersionKind:     gvk,
			groupVersionResource: restMapping.Resource,
			name:                 obj.GetName(),
			namespace:            namespace,
			obj:                  obj,
			raw:                  string(raw),
			namespaced:           namespaced,
		})
	}

	return currentKubernetesResources, nil
}
