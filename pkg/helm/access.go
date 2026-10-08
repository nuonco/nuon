package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// CheckResourceAccess lists each rendered kind in the namespace the chart will
// use. A forbidden response fails the deploy before Helm mutates the cluster.
// A missing namespace or a CRD that is not installed yet is skipped: Helm
// creates those as part of the release.
func CheckResourceAccess(ctx context.Context, kubeCfg *rest.Config, resources []Resource, releaseNamespace string, l *zap.Logger) error {
	if len(resources) == 0 {
		return nil
	}

	dyn, err := dynamic.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes client: %w", err)
	}
	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("unable to create kubernetes client: %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client.Discovery()))

	return checkResourceAccess(ctx, mapper, func(ctx context.Context, gvr schema.GroupVersionResource, namespace string) error {
		opts := metav1.ListOptions{Limit: 1}
		if namespace == "" {
			_, err := dyn.Resource(gvr).List(ctx, opts)
			return err
		}
		_, err := dyn.Resource(gvr).Namespace(namespace).List(ctx, opts)
		return err
	}, resources, releaseNamespace, l)
}

func checkResourceAccess(
	ctx context.Context,
	mapper meta.RESTMapper,
	list func(context.Context, schema.GroupVersionResource, string) error,
	resources []Resource,
	releaseNamespace string,
	l *zap.Logger,
) error {
	seen := map[string]struct{}{}
	checked := 0
	for _, resource := range resources {
		gv, err := schema.ParseGroupVersion(resource.APIVersion)
		if err != nil {
			return fmt.Errorf("unable to parse apiVersion %q for %s %s: %w", resource.APIVersion, resource.Kind, resource.Name, err)
		}
		mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gv.Group, Kind: resource.Kind}, gv.Version)
		if err != nil {
			// The chart is about to install this CRD, so the kind is not served yet.
			if meta.IsNoMatchError(err) {
				l.Debug("skipping access check for unknown kind", zap.String("kind", resource.Kind), zap.String("apiVersion", resource.APIVersion))
				continue
			}
			return fmt.Errorf("unable to map %s %s: %w", resource.Kind, resource.Name, err)
		}

		namespace := ""
		if mapping.Scope == nil || mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			namespace = resource.namespaceOr(releaseNamespace)
		}

		key := mapping.Resource.String() + "\x00" + namespace
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		l.Debug("checking chart resource access", zap.String("resource", mapping.Resource.Resource), zap.String("namespace", namespace))
		if err := list(ctx, mapping.Resource, namespace); err != nil {
			// The namespace or the resource itself may not exist until Helm creates it.
			if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
				l.Debug("skipping access check for missing resource", zap.String("resource", mapping.Resource.Resource), zap.String("namespace", namespace))
				continue
			}
			if namespace == "" {
				return fmt.Errorf("unable to list cluster-scoped %s: %w", mapping.Resource.Resource, err)
			}
			return fmt.Errorf("unable to list %s in namespace %q: %w", mapping.Resource.Resource, namespace, err)
		}
		checked++
	}

	l.Info("checked chart resource access", zap.Int("kinds", checked))
	return nil
}
