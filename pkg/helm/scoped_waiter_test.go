package helm

import (
	"context"
	"testing"

	"github.com/fluxcd/cli-utils/pkg/kstatus/polling/event"
	"github.com/fluxcd/cli-utils/pkg/kstatus/status"
	"github.com/fluxcd/cli-utils/pkg/kstatus/watcher"
	"github.com/fluxcd/cli-utils/pkg/object"
	"helm.sh/helm/v4/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestWaitListsServicesInReleaseAndObservedNamespaces(t *testing.T) {
	paused := true
	resources := kubeResourceList(
		serviceInfo("whoami", ""),
		serviceInfo("edge", "sourdough"),
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

	capture := &captureWatcher{}
	waiter := &scopedStatusWaiter{namespace: "persimmon"}
	if err := waiter.wait(context.Background(), resources, capture, true, status.CurrentStatus); err != nil {
		t.Fatalf("wait: %v", err)
	}

	if capture.opts.RESTScopeStrategy != watcher.RESTScopeNamespace {
		t.Fatalf("scope strategy = %v, want namespace", capture.opts.RESTScopeStrategy)
	}

	got := map[string]string{}
	for _, id := range capture.ids {
		got[id.GroupKind.Kind+"/"+id.Name] = id.Namespace
	}
	want := map[string]string{
		"Service/whoami":    "persimmon",
		"Service/edge":      "sourdough",
		"ClusterRole/admin": "",
	}
	if len(got) != len(want) {
		t.Fatalf("watched %#v, want %#v", got, want)
	}
	for key, namespace := range want {
		if got[key] != namespace {
			t.Fatalf("%s namespace = %q, want %q", key, got[key], namespace)
		}
	}
	if resources[0].Namespace != "persimmon" {
		t.Fatalf("service info namespace = %q, want release namespace", resources[0].Namespace)
	}
}

type captureWatcher struct {
	ids  object.ObjMetadataSet
	opts watcher.Options
}

func (c *captureWatcher) Watch(_ context.Context, ids object.ObjMetadataSet, opts watcher.Options) <-chan event.Event {
	c.ids = append(object.ObjMetadataSet{}, ids...)
	c.opts = opts
	ch := make(chan event.Event)
	close(ch)
	return ch
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

func kubeResourceList(infos ...*resource.Info) kube.ResourceList {
	return infos
}
