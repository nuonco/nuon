package componenthealth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
