package helm

import (
	"testing"
	"time"

	"helm.sh/helm/v4/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestGetWaiterWithOptionsStampsNamespaces(t *testing.T) {
	paused := true
	resources := kubeResourceList(
		serviceInfo("whoami", ""),
		serviceInfo("edge", "sourdough"),
		secretInfo("token", "rye"),
		&resource.Info{
			Name: "paused",
			Mapping: &meta.RESTMapping{
				GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
				Scope:            meta.RESTScopeNamespace,
			},
			Object: &appsv1.Deployment{
				TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
				ObjectMeta: metav1.ObjectMeta{Name: "paused", Namespace: "persimmon"},
				Spec:       appsv1.DeploymentSpec{Paused: paused},
			},
		},
		&resource.Info{
			Name: "admin",
			Mapping: &meta.RESTMapping{
				GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
				Scope:            meta.RESTScopeRoot,
			},
			Object: &rbacv1.ClusterRole{
				TypeMeta:   metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
				ObjectMeta: metav1.ObjectMeta{Name: "admin"},
			},
		},
	)

	// GetWaiter and GetWaiterWithOptions both return this waiter around Helm's.
	inner := &recordingWaiter{}
	waiter, err := wrapNamespaceWaiter(inner, nil, "persimmon")
	if err != nil {
		t.Fatalf("wrap waiter: %v", err)
	}
	if err := waiter.Wait(resources, time.Second); err != nil {
		t.Fatalf("wait: %v", err)
	}

	if got := resources[0].Namespace; got != "persimmon" {
		t.Fatalf("omitted service namespace = %q, want release namespace", got)
	}
	if got := resources[0].Object.(*corev1.Service).Namespace; got != "persimmon" {
		t.Fatalf("omitted service object namespace = %q, want release namespace", got)
	}
	if got := resources[1].Object.(*corev1.Service).Namespace; got != "sourdough" {
		t.Fatalf("observed service namespace = %q, want sourdough", got)
	}
	if got := resources[2].Object.(*corev1.Secret).Namespace; got != "rye" {
		t.Fatalf("secret namespace = %q, want rye", got)
	}
	deployment := resources[3].Object.(*appsv1.Deployment)
	if deployment.Namespace != "persimmon" || !deployment.Spec.Paused {
		t.Fatalf("paused deployment = namespace %q paused %v, want persimmon and still paused", deployment.Namespace, deployment.Spec.Paused)
	}
	if got := resources[4].Object.(*rbacv1.ClusterRole).Namespace; got != "" {
		t.Fatalf("cluster role namespace = %q, want empty", got)
	}
	if inner.resources[2].Object.(*corev1.Secret).Namespace != "rye" {
		t.Fatalf("helm waiter saw secret namespace %q, want rye", inner.resources[2].Object.(*corev1.Secret).Namespace)
	}
}

type recordingWaiter struct {
	resources kube.ResourceList
}

func (w *recordingWaiter) Wait(resources kube.ResourceList, _ time.Duration) error {
	w.resources = resources
	return nil
}

func (w *recordingWaiter) WaitWithJobs(resources kube.ResourceList, timeout time.Duration) error {
	return w.Wait(resources, timeout)
}

func (w *recordingWaiter) WaitForDelete(resources kube.ResourceList, timeout time.Duration) error {
	return w.Wait(resources, timeout)
}

func (w *recordingWaiter) WatchUntilReady(resources kube.ResourceList, timeout time.Duration) error {
	return w.Wait(resources, timeout)
}

func serviceInfo(name, namespace string) *resource.Info {
	return &resource.Info{
		Name:      name,
		Namespace: namespace,
		Mapping: &meta.RESTMapping{
			GroupVersionKind: schema.GroupVersionKind{Version: "v1", Kind: "Service"},
			Scope:            meta.RESTScopeNamespace,
		},
		Object: &corev1.Service{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		},
	}
}

func secretInfo(name, namespace string) *resource.Info {
	return &resource.Info{
		Name:      name,
		Namespace: namespace,
		Mapping: &meta.RESTMapping{
			GroupVersionKind: schema.GroupVersionKind{Version: "v1", Kind: "Secret"},
			Scope:            meta.RESTScopeNamespace,
		},
		Object: &corev1.Secret{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
	}
}

func kubeResourceList(infos ...*resource.Info) kube.ResourceList {
	return infos
}
