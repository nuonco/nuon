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

func TestRequiredKind(t *testing.T) {
	statefulsets := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	pods := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	nodePools := schema.GroupVersionResource{Group: "karpenter.sh", Version: "v1", Resource: "nodepools"}

	p := NewManifestKindsProvider(ManifestKindsProviderParams{})
	e := &Engine{manifestKinds: p}

	assert.True(t, e.requiredKind(statefulsets), "nothing recorded yet means nothing can be ruled out")

	p.Set("cmp-a", "---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api\n", "app")
	assert.True(t, e.requiredKind(deploymentsGVR), "a rendered kind is required")
	assert.False(t, e.requiredKind(statefulsets), "a core kind nothing renders is not")
	assert.True(t, e.requiredKind(pods), "pods are always required")
	assert.True(t, e.requiredKind(nodePools), "discovered kinds are rendered by definition")
}

func TestForgivePodlessNamespaces(t *testing.T) {
	forbidden := func(ns string) error {
		return apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, ns, nil)
	}
	helmDeployment := func(name, ns, release string) *unstructured.Unstructured {
		u := controller("Deployment", name, ns, "", "")
		u.SetLabels(map[string]string{helmManagedByLabel: helmManagedByValue})
		u.SetAnnotations(map[string]string{helmReleaseNameAnnotation: release})
		return u
	}
	listErr := func() error {
		return &namespaceListError{failed: map[string]error{
			"control": forbidden("control"),
			"apps":    forbidden("apps"),
			"sandbox": forbidden("sandbox"),
		}}
	}
	objects := []listedObject{
		{gvr: deploymentsGVR, u: helmDeployment("proxy", "control", "runtime")},
		{gvr: deploymentsGVR, u: helmDeployment("customer-app", "apps", "someone-else")},
	}

	newEngine := func(manifest string) *Engine {
		kinds := NewManifestKindsProvider(ManifestKindsProviderParams{})
		kinds.Set("cmp-1", manifest, "control")
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

	t.Run("keeps only namespaces running a component workload", func(t *testing.T) {
		e := newEngine("---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: proxy\n")
		err := e.forgivePodlessNamespaces("ins-1", listErr(), objects)

		var nsErr *namespaceListError
		require.ErrorAs(t, err, &nsErr)
		assert.Equal(t, []string{"control"}, nsErr.namespaces())
		assert.True(t, apierrors.IsForbidden(err))
	})

	t.Run("nothing left is no error", func(t *testing.T) {
		e := newEngine("---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: proxy\n")
		err := e.forgivePodlessNamespaces("ins-1", listErr(), objects[1:])
		assert.NoError(t, err)
	})

	t.Run("a rendered bare pod keeps every namespace", func(t *testing.T) {
		e := newEngine("---\napiVersion: v1\nkind: Pod\nmetadata:\n  name: once\n")
		err := e.forgivePodlessNamespaces("ins-1", listErr(), objects)

		var nsErr *namespaceListError
		require.ErrorAs(t, err, &nsErr)
		assert.Len(t, nsErr.namespaces(), 3)
	})
}
