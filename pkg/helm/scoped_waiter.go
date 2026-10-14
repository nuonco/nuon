package helm

import (
	"time"

	"helm.sh/helm/v4/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
)

// Helm 4.1 lists each kind in the object's own namespace. An object that
// omitted metadata.namespace is still listed at cluster scope, because an
// empty namespace is a cluster list. This client fills that namespace in
// before Helm's waiter runs: the namespace already on the object, then the
// namespace Helm applied with, then the release namespace.
type scopedKubeClient struct {
	*kube.Client
	namespace string
}

func newScopedKubeClient(client *kube.Client, namespace string) *scopedKubeClient {
	return &scopedKubeClient{Client: client, namespace: namespace}
}

// Helm install and upgrade call GetWaiterWithOptions when the client
// implements it. Both entry points have to stamp namespaces; the embedded
// client would otherwise satisfy the interface and skip this wrapper.
func (c *scopedKubeClient) GetWaiter(strategy kube.WaitStrategy) (kube.Waiter, error) {
	return c.GetWaiterWithOptions(strategy)
}

func (c *scopedKubeClient) GetWaiterWithOptions(strategy kube.WaitStrategy, opts ...kube.WaitOption) (kube.Waiter, error) {
	waiter, err := c.Client.GetWaiterWithOptions(strategy, opts...)
	return wrapNamespaceWaiter(waiter, err, c.namespace)
}

var _ kube.InterfaceWaitOptions = (*scopedKubeClient)(nil)

func wrapNamespaceWaiter(waiter kube.Waiter, err error, namespace string) (kube.Waiter, error) {
	if err != nil {
		return nil, err
	}
	return &namespaceStampWaiter{inner: waiter, namespace: namespace}, nil
}

type namespaceStampWaiter struct {
	inner     kube.Waiter
	namespace string
}

func (w *namespaceStampWaiter) Wait(resources kube.ResourceList, timeout time.Duration) error {
	stampReleaseNamespace(resources, w.namespace)
	return w.inner.Wait(resources, timeout)
}

func (w *namespaceStampWaiter) WaitWithJobs(resources kube.ResourceList, timeout time.Duration) error {
	stampReleaseNamespace(resources, w.namespace)
	return w.inner.WaitWithJobs(resources, timeout)
}

func (w *namespaceStampWaiter) WaitForDelete(resources kube.ResourceList, timeout time.Duration) error {
	stampReleaseNamespace(resources, w.namespace)
	return w.inner.WaitForDelete(resources, timeout)
}

func (w *namespaceStampWaiter) WatchUntilReady(resources kube.ResourceList, timeout time.Duration) error {
	stampReleaseNamespace(resources, w.namespace)
	return w.inner.WatchUntilReady(resources, timeout)
}

func stampReleaseNamespace(resources kube.ResourceList, releaseNamespace string) {
	for _, info := range resources {
		if info == nil || info.Object == nil || !namespacedInfo(info) {
			continue
		}
		accessor, err := meta.Accessor(info.Object)
		if err != nil {
			continue
		}
		namespace := namespaceForWait(accessor.GetNamespace(), info, releaseNamespace)
		if accessor.GetNamespace() != namespace {
			accessor.SetNamespace(namespace)
		}
		info.Namespace = namespace
	}
}

func namespaceForWait(objNamespace string, info *resource.Info, releaseNamespace string) string {
	if objNamespace != "" {
		return objNamespace
	}
	if info != nil && info.Namespace != "" {
		return info.Namespace
	}
	if releaseNamespace != "" {
		return releaseNamespace
	}
	return corev1.NamespaceDefault
}

func namespacedInfo(info *resource.Info) bool {
	return info != nil && info.Mapping != nil && info.Mapping.Scope != nil && info.Mapping.Scope.Name() == meta.RESTScopeNameNamespace
}
