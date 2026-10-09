package componenthealth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
)

var deploymentsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

// denyClusterWide mimics a Role-only identity: list is allowed in the given
// namespaces and forbidden everywhere else, cluster-wide included.
func denyClusterWide(allowed ...string) k8stesting.ReactionFunc {
	ok := map[string]struct{}{}
	for _, ns := range allowed {
		ok[ns] = struct{}{}
	}
	return func(action k8stesting.Action) (bool, runtime.Object, error) {
		if _, allowed := ok[action.GetNamespace()]; allowed {
			return false, nil, nil
		}
		gr := schema.GroupResource{Group: action.GetResource().Group, Resource: action.GetResource().Resource}
		return true, nil, apierrors.NewForbidden(gr, "", nil)
	}
}

func TestListScoped(t *testing.T) {
	ctx := context.Background()
	objs := []runtime.Object{
		controller("Deployment", "api", "app", "", ""),
		controller("Deployment", "worker", "jobs", "", ""),
		controller("Deployment", "coredns", "kube-system", "", ""),
	}

	t.Run("cluster-wide list is used when allowed", func(t *testing.T) {
		items, err := listScoped(ctx, fakeDyn(objs...), deploymentsGVR, metav1.ListOptions{}, []string{"app"})
		require.NoError(t, err)
		assert.Len(t, items, 3)
	})

	t.Run("forbidden cluster-wide falls back to known namespaces", func(t *testing.T) {
		dyn := fakeDyn(objs...)
		dyn.PrependReactor("list", "deployments", denyClusterWide("app", "jobs"))

		items, err := listScoped(ctx, dyn, deploymentsGVR, metav1.ListOptions{}, []string{"app", "jobs"})
		require.NoError(t, err)
		names := []string{}
		for _, u := range items {
			names = append(names, u.GetName())
		}
		assert.ElementsMatch(t, []string{"api", "worker"}, names)
	})

	t.Run("a forbidden namespace keeps the rest and reports the error", func(t *testing.T) {
		dyn := fakeDyn(objs...)
		dyn.PrependReactor("list", "deployments", denyClusterWide("app"))

		items, err := listScoped(ctx, dyn, deploymentsGVR, metav1.ListOptions{}, []string{"app", "jobs"})
		require.Error(t, err)
		assert.True(t, apierrors.IsForbidden(err))
		require.Len(t, items, 1)
		assert.Equal(t, "api", items[0].GetName())
	})

	t.Run("no known namespaces surfaces the forbidden error", func(t *testing.T) {
		dyn := fakeDyn(objs...)
		dyn.PrependReactor("list", "deployments", denyClusterWide())

		items, err := listScoped(ctx, dyn, deploymentsGVR, metav1.ListOptions{}, nil)
		require.Error(t, err)
		assert.Empty(t, items)
	})
}

func TestManifestKindsNamespaces(t *testing.T) {
	p := NewManifestKindsProvider(ManifestKindsProviderParams{})

	p.Set("cmp-a", "---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api\n---\napiVersion: v1\nkind: Service\nmetadata:\n  name: api\n  namespace: edge\n", "app")
	p.SetKinds("cmp-tf", nil, []string{resourceKey("Deployment", "tf-ns", "x")}, nil)
	assert.Equal(t, []string{"app", "edge", "tf-ns"}, p.Namespaces())

	p.Set("cmp-a", "", "app")
	assert.Equal(t, []string{"tf-ns"}, p.Namespaces(), "an empty manifest clears the component's namespaces")
}

func TestForgiveForbidden(t *testing.T) {
	forbidden := func(resource, ns string) error {
		return apierrors.NewForbidden(schema.GroupResource{Resource: resource}, ns, nil)
	}
	listErr := func(resource string, namespaces ...string) error {
		failed := map[string]error{}
		for _, ns := range namespaces {
			failed[ns] = forbidden(resource, ns)
		}
		return &namespaceListError{failed: failed}
	}
	failedNamespaces := func(t *testing.T, err error) []string {
		if err == nil {
			return nil
		}
		var nsErr *namespaceListError
		require.ErrorAs(t, err, &nsErr)
		return nsErr.namespaces()
	}
	helmDeployment := func(name, ns, release string) *unstructured.Unstructured {
		u := controller("Deployment", name, ns, "", "")
		u.SetLabels(map[string]string{helmManagedByLabel: helmManagedByValue})
		u.SetAnnotations(map[string]string{helmReleaseNameAnnotation: release})
		return u
	}

	// A chart running a Deployment in control and only dropping a
	// ServiceAccount into apps, whose Role grants no workload reads.
	const chart = "---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: proxy\n---\napiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: storage\n  namespace: apps\n"
	newEngine := func(manifest string) *Engine {
		kinds := NewManifestKindsProvider(ManifestKindsProviderParams{})
		if manifest != "" {
			kinds.Set("cmp-1", manifest, "control")
		}
		e := &Engine{
			l:             zap.NewNop(),
			idx:           newIndex(),
			cluster:       NewClusterProvider(ProviderParams{L: zap.NewNop()}),
			manifestKinds: kinds,
		}
		e.idx.replace("ins-1", []componentEntry{
			{installComponentID: "ic-1", componentID: "cmp-1", componentType: "helm_chart", helmReleaseName: "runtime"},
		})
		return e
	}
	objects := []listedObject{
		{gvr: deploymentsGVR, u: helmDeployment("proxy", "control", "runtime")},
		{gvr: deploymentsGVR, u: helmDeployment("customer-app", "apps", "someone-else")},
	}
	ingresses := schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}

	t.Run("a kind is only a gap where a component renders it", func(t *testing.T) {
		e := newEngine(chart)
		err := e.forgiveForbidden("ins-1", deploymentsGVR, coreKinds[deploymentsGVR], listErr("deployments", "control", "apps"), objects)
		assert.Equal(t, []string{"control"}, failedNamespaces(t, err))
		assert.True(t, apierrors.IsForbidden(err))

		err = e.forgiveForbidden("ins-1", ingresses, coreKinds[ingresses], listErr("ingresses", "control", "apps"), objects)
		assert.NoError(t, err, "nothing renders an Ingress")
	})

	t.Run("pods are a gap only where a component workload runs", func(t *testing.T) {
		e := newEngine(chart)
		err := e.forgiveForbidden("ins-1", podsGVR, coreKinds[podsGVR], listErr("pods", "control", "apps"), objects)
		assert.Equal(t, []string{"control"}, failedNamespaces(t, err))

		err = e.forgiveForbidden("ins-1", podsGVR, coreKinds[podsGVR], listErr("pods", "control", "apps"), objects[1:])
		assert.NoError(t, err)
	})

	t.Run("a rendered bare pod keeps its namespace", func(t *testing.T) {
		e := newEngine("---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cfg\n---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: once\n  namespace: apps\n")
		err := e.forgiveForbidden("ins-1", podsGVR, coreKinds[podsGVR], listErr("pods", "control", "apps"), nil)
		assert.Equal(t, []string{"apps"}, failedNamespaces(t, err))
	})

	t.Run("identity kinds are never a gap", func(t *testing.T) {
		e := newEngine(chart)
		namespaces := schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
		err := e.forgiveForbidden("ins-1", namespaces, schema.GroupKind{Kind: "Namespace"}, forbidden("namespaces", ""), objects)
		assert.NoError(t, err)
	})

	t.Run("nothing recorded rules nothing out", func(t *testing.T) {
		e := newEngine("")
		err := e.forgiveForbidden("ins-1", ingresses, coreKinds[ingresses], listErr("ingresses", "control", "apps"), objects)
		assert.Equal(t, []string{"apps", "control"}, failedNamespaces(t, err))
	})

	t.Run("a namespace no recorded component covers rules nothing out", func(t *testing.T) {
		e := newEngine(chart)
		err := e.forgiveForbidden("ins-1", ingresses, coreKinds[ingresses], listErr("ingresses", "legacy"), objects)
		assert.Equal(t, []string{"legacy"}, failedNamespaces(t, err))
	})
}

func TestManifestKindsRendersInSurvivesRestart(t *testing.T) {
	cluster := NewClusterProvider(ProviderParams{L: zap.NewNop()})
	p := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: cluster})
	p.Set("cmp-1", "---\napiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: web\n---\napiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: sa\n  namespace: apps\n", "control")

	restored := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: cluster})
	restored.Load()
	ingress := schema.GroupKind{Group: "networking.k8s.io", Kind: "Ingress"}
	assert.True(t, restored.RendersIn("control", ingress))
	assert.False(t, restored.RendersIn("apps", ingress))
}

func TestManifestKindsRendersInLegacyEntries(t *testing.T) {
	cluster := NewClusterProvider(ProviderParams{L: zap.NewNop()})
	cluster.SetComponentKinds([]string{"cmp-1|apps/v1/Deployment", namespaceEntryPrefix + "cmp-1|web"})
	p := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: cluster})
	p.Load()

	assert.True(t, p.RendersIn("web", schema.GroupKind{Group: "apps", Kind: "Deployment"}))
	assert.False(t, p.RendersIn("web", schema.GroupKind{Group: "apps", Kind: "StatefulSet"}))
}
