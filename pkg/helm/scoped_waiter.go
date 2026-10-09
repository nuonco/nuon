package helm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/fluxcd/cli-utils/pkg/kstatus/polling/aggregator"
	"github.com/fluxcd/cli-utils/pkg/kstatus/polling/collector"
	"github.com/fluxcd/cli-utils/pkg/kstatus/polling/event"
	"github.com/fluxcd/cli-utils/pkg/kstatus/status"
	"github.com/fluxcd/cli-utils/pkg/kstatus/watcher"
	"github.com/fluxcd/cli-utils/pkg/object"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"helm.sh/helm/v4/pkg/kube"
)

// Helm's status watcher lists each kind at cluster scope when the release
// contains objects in more than one namespace. A namespace Role cannot satisfy
// that list. This client keeps the watcher, but lists each kind in the object's
// namespace. A namespaced object that omitted metadata.namespace uses the
// resource namespace Helm already filled in, then the release namespace.
type scopedKubeClient struct {
	*kube.Client
	namespace string
}

func newScopedKubeClient(client *kube.Client, namespace string) *scopedKubeClient {
	return &scopedKubeClient{Client: client, namespace: namespace}
}

func (c *scopedKubeClient) GetWaiter(strategy kube.WaitStrategy) (kube.Waiter, error) {
	switch strategy {
	case kube.LegacyStrategy:
		return c.Client.GetWaiter(strategy)
	default:
		waiter, err := c.newStatusWaiter()
		if err != nil {
			return nil, err
		}
		if strategy == kube.HookOnlyStrategy {
			return &hookOnlyScopedWaiter{watch: waiter}, nil
		}
		return waiter, nil
	}
}

func (c *scopedKubeClient) newStatusWaiter() (*scopedStatusWaiter, error) {
	cfg, err := c.Factory.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dynamicClient, err := c.Factory.DynamicClient()
	if err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	restMapper, err := apiutil.NewDynamicRESTMapper(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &scopedStatusWaiter{
		kube:       c.Client,
		dynamic:    dynamicClient,
		restMapper: restMapper,
		namespace:  c.namespace,
	}, nil
}

type scopedStatusWaiter struct {
	kube       *kube.Client
	dynamic    dynamic.Interface
	restMapper meta.RESTMapper
	namespace  string
}

func (w *scopedStatusWaiter) Wait(resources kube.ResourceList, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	return w.wait(ctx, resources, w.newWatcher(), true, status.CurrentStatus)
}

func (w *scopedStatusWaiter) WaitWithJobs(resources kube.ResourceList, timeout time.Duration) error {
	return w.Wait(resources, timeout)
}

func (w *scopedStatusWaiter) WaitForDelete(resources kube.ResourceList, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()
	return w.wait(ctx, resources, w.newWatcher(), false, status.NotFoundStatus)
}

func (w *scopedStatusWaiter) WatchUntilReady(resources kube.ResourceList, timeout time.Duration) error {
	if _, err := prepareWaitResources(resources, w.namespace, false); err != nil {
		return err
	}
	legacy, err := w.kube.GetWaiter(kube.LegacyStrategy)
	if err != nil {
		return err
	}
	return legacy.WatchUntilReady(resources, timeout)
}

func (w *scopedStatusWaiter) newWatcher() watcher.StatusWatcher {
	return watcher.NewDefaultStatusWatcher(w.dynamic, w.restMapper)
}

func (w *scopedStatusWaiter) wait(ctx context.Context, resources kube.ResourceList, sw watcher.StatusWatcher, skipPaused bool, desired status.Status) error {
	ids, err := prepareWaitResources(resources, w.namespace, skipPaused)
	if err != nil {
		return err
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	eventCh := sw.Watch(cancelCtx, ids, watcher.Options{RESTScopeStrategy: watcher.RESTScopeNamespace})
	statusCollector := collector.NewResourceStatusCollector(ids)
	done := statusCollector.ListenWithObserver(eventCh, statusObserver(cancel, desired))
	<-done

	if statusCollector.Error != nil {
		return statusCollector.Error
	}
	if ctx.Err() == nil {
		return nil
	}

	errs := make([]error, 0, len(ids)+1)
	for _, id := range ids {
		rs := statusCollector.ResourceStatuses[id]
		if rs == nil || rs.Status == desired {
			continue
		}
		errs = append(errs, fmt.Errorf("resource not ready, name: %s, kind: %s, status: %s", rs.Identifier.Name, rs.Identifier.GroupKind.Kind, rs.Status))
	}
	errs = append(errs, ctx.Err())
	return errors.Join(errs...)
}

func statusObserver(cancel context.CancelFunc, desired status.Status) collector.ObserverFunc {
	return func(statusCollector *collector.ResourceStatusCollector, _ event.Event) {
		var statuses []*event.ResourceStatus
		var pending []*event.ResourceStatus
		for _, rs := range statusCollector.ResourceStatuses {
			if rs == nil {
				continue
			}
			if rs.Status == status.UnknownStatus && desired == status.NotFoundStatus {
				continue
			}
			statuses = append(statuses, rs)
			if rs.Status != desired {
				pending = append(pending, rs)
			}
		}
		if aggregator.AggregateStatus(statuses, desired) == desired {
			cancel()
			return
		}
		if len(pending) == 0 {
			return
		}
		sort.Slice(pending, func(i, j int) bool {
			return pending[i].Identifier.Name < pending[j].Identifier.Name
		})
		first := pending[0]
		slog.Debug("waiting for resource", "name", first.Identifier.Name, "kind", first.Identifier.GroupKind.Kind, "expectedStatus", desired, "actualStatus", first.Status)
	}
}

type hookOnlyScopedWaiter struct {
	watch kube.Waiter
}

func (w *hookOnlyScopedWaiter) Wait(kube.ResourceList, time.Duration) error { return nil }

func (w *hookOnlyScopedWaiter) WaitWithJobs(kube.ResourceList, time.Duration) error { return nil }

func (w *hookOnlyScopedWaiter) WaitForDelete(kube.ResourceList, time.Duration) error {
	return nil
}

func (w *hookOnlyScopedWaiter) WatchUntilReady(resources kube.ResourceList, timeout time.Duration) error {
	return w.watch.WatchUntilReady(resources, timeout)
}

func prepareWaitResources(resources kube.ResourceList, releaseNamespace string, skipPaused bool) ([]object.ObjMetadata, error) {
	ids := make([]object.ObjMetadata, 0, len(resources))
	for _, info := range resources {
		if info == nil || info.Object == nil {
			continue
		}
		if skipPaused && pausedDeployment(info) {
			continue
		}
		obj, err := object.RuntimeToObjMeta(info.Object)
		if err != nil {
			return nil, err
		}
		obj.Namespace = namespaceForWait(obj.Namespace, info, releaseNamespace)
		if namespacedInfo(info) {
			info.Namespace = obj.Namespace
		}
		ids = append(ids, obj)
	}
	return ids, nil
}

func namespaceForWait(objNamespace string, info *resource.Info, releaseNamespace string) string {
	if !namespacedInfo(info) {
		return ""
	}
	if objNamespace != "" {
		return objNamespace
	}
	if info.Namespace != "" {
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

func pausedDeployment(info *resource.Info) bool {
	deployment, ok := kube.AsVersioned(info).(*appsv1.Deployment)
	return ok && deployment.Spec.Paused
}
