package componenthealth

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func liftPodHealthToOwnersIn(byKey map[string]*unstructured.Unstructured) map[string]string {
	return liftPodHealthToOwners(failedPods(byKey), byKey)
}
