package componenthealth

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type resourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	UID        string
}

func resourceRefForObject(u *unstructured.Unstructured) resourceRef {
	if u == nil {
		return resourceRef{}
	}
	return resourceRef{
		APIVersion: u.GetAPIVersion(),
		Kind:       u.GetKind(),
		Namespace:  u.GetNamespace(),
		Name:       u.GetName(),
		UID:        string(u.GetUID()),
	}
}

func resourceRefForOwner(ref *metav1.OwnerReference, namespace string) resourceRef {
	if ref == nil {
		return resourceRef{}
	}
	return resourceRef{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Namespace:  namespace,
		Name:       ref.Name,
		UID:        string(ref.UID),
	}
}

func (r resourceRef) valid() bool {
	return r.Kind != "" && r.Name != ""
}

func (r resourceRef) key() string {
	return r.APIVersion + "\x00" + r.Kind + "\x00" + r.Namespace + "\x00" + r.Name
}

func (r resourceRef) matches(u *unstructured.Unstructured) bool {
	if !r.valid() || u == nil || r.key() != resourceRefForObject(u).key() {
		return false
	}
	return r.UID == "" || string(u.GetUID()) == r.UID
}

func (r resourceRef) details() map[string]any {
	out := map[string]any{
		"api_version": r.APIVersion,
		"kind":        r.Kind,
		"name":        r.Name,
	}
	if r.Namespace != "" {
		out["namespace"] = r.Namespace
	}
	if r.UID != "" {
		out["uid"] = r.UID
	}
	return out
}
