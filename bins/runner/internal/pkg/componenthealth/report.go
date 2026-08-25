package componenthealth

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/conc"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	providerKubernetes = "kubernetes"

	// maxResourcesPerComponent mirrors the server-side cap on a single
	// component's resource list, bounded here so the server never truncates.
	maxResourcesPerComponent = 500
)

// report runs one stateless cycle: refresh the index, collect resources from
// every surface (cluster objects, probes, terraform state), and push one batch.
func (e *Engine) report(ctx context.Context) error {
	if err := e.rebuildIndex(ctx); err != nil {
		return fmt.Errorf("unable to refresh component index: %w", err)
	}
	installID := e.idx.installIDValue()
	if installID == "" {
		return nil
	}

	grouped := map[string][]*models.ServiceComponentHealthResource{}
	sandbox := map[string][]*models.ServiceComponentHealthResource{}
	sandboxNS := map[string]string{}

	e.collectProbes(ctx, grouped)

	// Missing or transiently broken cluster access is not a reason to drop the
	// surfaces that did report.
	//
	// Cluster access is one fact about the install, so it is reported once at
	// install level instead of copied onto every component. A component that
	// cannot be inspected simply reports nothing this cycle and ages into
	// unknown through the normal staleness path, which also clears it on
	// recovery — a per-component row would linger forever instead.
	var clusterErr string
	if ci := e.cluster.Get(); ci == nil {
		e.l.Info("component health: waiting for cluster access (not yet provided by a deploy)")
		clusterErr = "no cluster access stored for this install yet; refresh it from the install's health card"
	} else if err := e.collectCluster(ctx, installID, ci, grouped, sandbox, sandboxNS, &clusterErr); err != nil {
		e.l.Error("unable to collect cluster resources for component health", zap.Error(err))
		clusterErr = fmt.Sprintf("unable to inspect cluster resources: %v", err)
	}

	e.collectTerraform(grouped)

	if len(grouped) == 0 && len(sandbox) == 0 {
		return nil
	}

	components := make([]*models.ServiceComponentHealthComponent, 0, len(grouped))
	for componentID, resources := range grouped {
		entry, ok := e.idx.lookup(componentID)
		if !ok {
			continue
		}
		installComponentID := entry.installComponentID
		bounded, truncated := e.bound(resources, componentID)
		components = append(components, &models.ServiceComponentHealthComponent{
			InstallComponentID: &installComponentID,
			ComponentID:        entry.componentID,
			ComponentType:      entry.componentType,
			Resources:          bounded,
			Truncated:          truncated,
		})
	}

	sandboxReleases := make([]*models.ServiceComponentHealthSandboxRelease, 0, len(sandbox))
	for release, resources := range sandbox {
		releaseName := release
		bounded, _ := e.bound(resources, release)
		sandboxReleases = append(sandboxReleases, &models.ServiceComponentHealthSandboxRelease{
			ReleaseName: &releaseName,
			Namespace:   sandboxNS[release],
			Resources:   bounded,
		})
	}

	req := &models.ServiceCreateComponentHealthRequest{
		InstallID:          &installID,
		Kind:               "resync",
		ObservedAt:         time.Now().UTC().Format(time.RFC3339),
		Components:         components,
		SandboxReleases:    sandboxReleases,
		ClusterAccessError: clusterErr,
	}
	if _, err := e.apiClient.CreateComponentHealth(ctx, req); err != nil {
		return fmt.Errorf("unable to report component health: %w", err)
	}

	return nil
}

func (e *Engine) bound(resources []*models.ServiceComponentHealthResource, owner string) ([]*models.ServiceComponentHealthResource, bool) {
	if len(resources) <= maxResourcesPerComponent {
		return resources, false
	}
	e.l.Warn("truncating component health resource list",
		zap.String("owner", owner),
		zap.Int("resources", len(resources)),
		zap.Int("cap", maxResourcesPerComponent),
	)
	return resources[:maxResourcesPerComponent], true
}

// collectCluster lists the watched kinds plus current warning events, assesses
// each object, and groups it under the component or sandbox release owning it.
func (e *Engine) collectCluster(
	ctx context.Context,
	installID string,
	ci *kube.ClusterInfo,
	grouped, sandbox map[string][]*models.ServiceComponentHealthResource,
	sandboxNS map[string]string,
	clusterErr *string,
) error {
	restCfg, err := kube.ConfigForCluster(ctx, ci)
	if err != nil {
		return fmt.Errorf("unable to build cluster config: %w", err)
	}
	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("unable to build dynamic client: %w", err)
	}

	warnings := e.latestWarnings(ctx, dynClient)

	type listedObject struct {
		gvr schema.GroupVersionResource
		u   *unstructured.Unstructured
	}
	objects := make([]listedObject, 0, 256)
	byKey := map[string]*unstructured.Unstructured{}

	// Counting failed lists is what tells "owns nothing of that kind" apart from
	// "we were not allowed to look" — both otherwise arrive as an empty report.
	var failedKinds []string
	var firstListErr error

	for _, gvr := range e.watchList(ctx, restCfg) {
		list, err := dynClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			e.l.Warn("unable to list resources for component health",
				zap.String("resource", gvr.String()), zap.Error(err))
			failedKinds = append(failedKinds, gvr.Resource)
			if firstListErr == nil {
				firstListErr = err
			}
			continue
		}
		for i := range list.Items {
			u := &list.Items[i]
			objects = append(objects, listedObject{gvr: gvr, u: u})
			byKey[resourceRefForObject(u).key()] = u
		}
	}

	if len(failedKinds) > 0 && len(objects) == 0 {
		return fmt.Errorf("unable to list any watched resource kind: %w", firstListErr)
	}
	if len(failedKinds) > 0 {
		// Partial loss still ships the kinds that did list; the message is what
		// keeps the gap visible, since otherwise the report looks complete.
		*clusterErr = fmt.Sprintf("unable to list %s: %v", strings.Join(failedKinds, ", "), firstListErr)
	}

	// The owner chain is walked through byKey, so intermediate owners that are
	// not listed have to be fetched before lifting — otherwise the walk stops
	// short and an ImagePullBackOff reads as benign progressing for 10m.
	e.hydrateWarningOwners(ctx, dynClient, warnings, byKey)

	attributeWarningsToOwners(warnings, byKey)

	for _, obj := range objects {
		gvr, u := obj.gvr, obj.u

		componentID, isComponent := e.componentFor(installID, u)
		var release string
		var isSandbox bool
		if !isComponent {
			release, isSandbox = e.sandboxReleaseFor(u)
		}
		if !isComponent && !isSandbox {
			continue
		}

		var warn *warningEvent
		if w, ok := warnings[resourceRefForObject(u).key()]; ok && warningTargetsObject(w, u) {
			warn = &w
		}
		res := resourceModel(gvr, u, warn)

		if isComponent {
			grouped[componentID] = append(grouped[componentID], res)
			continue
		}
		sandbox[release] = append(sandbox[release], res)
		if sandboxNS[release] == "" {
			sandboxNS[release] = u.GetAnnotations()[helmReleaseNamespaceAnnotation]
		}
	}

	return nil
}

// watchList is the core workload kinds plus any kind this install's components
// actually deploy. Without it a component shipping only CRs (a Karpenter
// NodePool, a ClickHouseInstallation) reports nothing at all, because nobody
// lists its kind.
//
// Kinds come from terraform state. Helm-rendered kinds cannot be discovered
// from the runner: the manifest lives in the release Secret and health's
// identity is deliberately denied secret reads, so a chart shipping only CRs
// still needs its kind reachable another way.
//
// Resolving kind to a listable resource needs the cluster's own discovery data,
// so an unresolvable kind is simply skipped rather than guessed at.
func (e *Engine) watchList(ctx context.Context, restCfg *rest.Config) []schema.GroupVersionResource {
	out := make([]schema.GroupVersionResource, 0, len(watchedGVRs)+8)
	out = append(out, watchedGVRs...)

	var discovered []schema.GroupVersionKind
	if e.manifestKinds != nil {
		discovered = e.manifestKinds.DiscoveredGVKs()
	}
	if len(discovered) == 0 {
		return out
	}

	disco, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		e.l.Warn("unable to build discovery client for dynamic kinds", zap.Error(err))
		return out
	}
	groups, err := restmapper.GetAPIGroupResources(disco)
	if err != nil {
		e.l.Warn("unable to fetch api group resources for dynamic kinds", zap.Error(err))
		return out
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groups)

	seen := map[schema.GroupVersionResource]struct{}{}
	for _, gvr := range out {
		seen[gvr] = struct{}{}
	}
	for _, gvk := range discovered {
		m, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			e.l.Debug("skipping kind with no rest mapping",
				zap.String("gvk", gvk.String()), zap.Error(err))
			continue
		}
		if _, dup := seen[m.Resource]; dup {
			continue
		}
		seen[m.Resource] = struct{}{}
		out = append(out, m.Resource)
	}

	return out
}

// maxOwnerHops bounds the ownerReferences walk so a cyclic chain cannot spin.
const maxOwnerHops = 4

const maxWarningHydrationGets = 100

// hydrateWarningOwners GETs an attributable warning's source and owner chain so
// byKey can resolve it without adding high-volume child kinds to every list.
func (e *Engine) hydrateWarningOwners(
	ctx context.Context,
	dynClient dynamic.Interface,
	warnings map[string]warningEvent,
	byKey map[string]*unstructured.Unstructured,
) {
	if len(warnings) == 0 {
		return
	}

	ordered := make([]warningEvent, 0, len(warnings))
	for _, warn := range warnings {
		ordered = append(ordered, warn)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].at.Equal(ordered[j].at) {
			return ordered[i].source.key() < ordered[j].source.key()
		}
		return ordered[i].at.After(ordered[j].at)
	})

	gets := 0
	limitReached := false
	for _, warn := range ordered {
		if !warningMayPropagate(warn) {
			continue
		}
		cur, ok := byKey[warn.source.key()]
		if !ok {
			if !warningSourceNeedsHydration(warn.source) {
				continue
			}
			if gets == maxWarningHydrationGets {
				limitReached = true
				break
			}
			gets++
			var err error
			cur, err = getResource(ctx, dynClient, warn.source)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					e.l.Warn("unable to fetch warning source",
						zap.String("kind", warn.source.Kind),
						zap.String("name", warn.source.Name),
						zap.Error(err))
				}
				continue
			}
			byKey[resourceRefForObject(cur).key()] = cur
		}
		if !warn.source.matches(cur) || !warningPropagates(cur, warn) {
			continue
		}

		for hop := 0; hop < maxOwnerHops; hop++ {
			ownerRef := resourceRefForOwner(metav1.GetControllerOf(cur), cur.GetNamespace())
			if !ownerRef.valid() {
				break
			}
			ownerKey := ownerRef.key()
			if next, ok := byKey[ownerKey]; ok {
				if !ownerRef.matches(next) {
					break
				}
				cur = next
				continue
			}
			if _, ok := resourceGVR(ownerRef); !ok {
				break
			}
			if gets == maxWarningHydrationGets {
				limitReached = true
				break
			}
			gets++
			owner, err := getResource(ctx, dynClient, ownerRef)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					e.l.Warn("unable to fetch owner for warning source",
						zap.String("kind", ownerRef.Kind), zap.String("name", ownerRef.Name), zap.Error(err))
				}
				break
			}
			if !ownerRef.matches(owner) {
				break
			}
			byKey[ownerKey] = owner
			cur = owner
		}
	}
	if limitReached {
		e.l.Warn("component health warning hydration reached its request limit",
			zap.Int("limit", maxWarningHydrationGets))
	}
}

func getResource(
	ctx context.Context,
	dynClient dynamic.Interface,
	ref resourceRef,
) (*unstructured.Unstructured, error) {
	gvr, ok := resourceGVR(ref)
	if !ok {
		return nil, fmt.Errorf("unable to resolve %s %s", ref.Kind, ref.APIVersion)
	}
	resource := dynClient.Resource(gvr)
	if ref.Namespace == "" {
		return resource.Get(ctx, ref.Name, metav1.GetOptions{})
	}
	return resource.Namespace(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
}

func resourceGVR(ref resourceRef) (schema.GroupVersionResource, bool) {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return schema.GroupVersionResource{}, false
	}
	gvk := gv.WithKind(ref.Kind)
	gvr, ok := ownerGVRs[gvk]
	return gvr, ok
}

func warningSourceNeedsHydration(ref resourceRef) bool {
	return ref.APIVersion == "apps/v1" && ref.Kind == "ReplicaSet"
}

func warningMayPropagate(warn warningEvent) bool {
	return warn.source.Kind == "Pod" || warn.reason == "FailedCreate"
}

func warningPropagates(source *unstructured.Unstructured, warn warningEvent) bool {
	if warn.reason == "FailedCreate" {
		return true
	}
	return source.GetKind() == "Pod" && !podReady(source)
}

func attributeWarningsToOwners(warnings map[string]warningEvent, byKey map[string]*unstructured.Unstructured) {
	if len(warnings) == 0 {
		return
	}

	lifted := map[string]warningEvent{}
	for _, warn := range warnings {
		u, ok := byKey[warn.source.key()]
		if !ok || !warn.source.matches(u) || !warningPropagates(u, warn) {
			continue
		}
		path := ownerPath(u, byKey)
		if len(path) < 2 {
			continue
		}
		warn.ownerPath = path
		targetKey := path[len(path)-1].key()
		if current, exists := lifted[targetKey]; !exists || warn.at.After(current.at) {
			lifted[targetKey] = warn
		}
	}

	for key, warn := range lifted {
		if current, seen := warnings[key]; seen && current.source.matches(byKey[key]) {
			continue
		}
		warnings[key] = warn
	}
}

func ownerPath(u *unstructured.Unstructured, byKey map[string]*unstructured.Unstructured) []resourceRef {
	cur := u
	path := []resourceRef{resourceRefForObject(cur)}
	for hop := 0; hop < maxOwnerHops; hop++ {
		ownerRef := resourceRefForOwner(metav1.GetControllerOf(cur), cur.GetNamespace())
		if !ownerRef.valid() {
			break
		}
		next, ok := byKey[ownerRef.key()]
		if !ok || next == cur || !ownerRef.matches(next) {
			break
		}
		cur = next
		path = append(path, resourceRefForObject(cur))
	}
	return path
}

func warningTargetsObject(warn warningEvent, u *unstructured.Unstructured) bool {
	target := warn.source
	if len(warn.ownerPath) > 0 {
		target = warn.ownerPath[len(warn.ownerPath)-1]
	}
	return target.matches(u)
}

func podReady(u *unstructured.Unstructured) bool {
	conds, ok, _ := unstructured.NestedSlice(u.Object, "status", "conditions")
	if !ok {
		return false
	}
	for _, c := range conds {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if cond["type"] == "Ready" {
			return cond["status"] == "True"
		}
	}
	return false
}

// collectTerraform adds identity-only rows from the state this process last
// applied; a component whose state it has not seen is skipped.
func (e *Engine) collectTerraform(grouped map[string][]*models.ServiceComponentHealthResource) {
	if e.terraform == nil {
		return
	}

	for _, componentID := range e.terraform.ComponentIDs() {
		resources := e.terraform.Resources(componentID)
		if len(resources) == 0 {
			continue
		}
		// A component the index no longer carries has been removed from the
		// install; its resources are not ours to report.
		if _, ok := e.idx.lookup(componentID); !ok {
			continue
		}
		grouped[componentID] = append(grouped[componentID], resources...)
	}

	for _, componentID := range e.idx.componentsOfType(componentTypeTerraformModule) {
		if len(grouped[componentID]) == 0 {
			e.l.Debug("no terraform state available for component this cycle",
				zap.String("component_id", componentID))
		}
	}
}

// collectProbes runs every declared probe once per cycle, one row each. Collected
// first so truncation can never drop a vendor's own assertion.
func (e *Engine) collectProbes(ctx context.Context, grouped map[string][]*models.ServiceComponentHealthResource) {
	targets := e.idx.probeTargets()
	if len(targets) == 0 {
		return
	}

	client := newProbeHTTPClient()
	defer client.CloseIdleConnections()

	var mu sync.Mutex
	wg := conc.NewWaitGroup()
	for componentID, specs := range targets {
		for _, spec := range specs {
			wg.Go(func() {
				row := probeResourceRow(spec, runProbe(ctx, client, spec))
				mu.Lock()
				defer mu.Unlock()
				grouped[componentID] = append(grouped[componentID], row)
			})
		}
	}
	if rec := wg.WaitAndRecover(); rec != nil {
		e.l.Error("recovered panic while probing component health", zap.Error(rec.AsError()))
	}
}

func resourceModel(gvr schema.GroupVersionResource, u *unstructured.Unstructured, warn *warningEvent) *models.ServiceComponentHealthResource {
	health, message, native := assessResource(u)

	// A current Warning event means the resource is failing even when its own
	// status looks benign (an Ingress stuck Progressing on a listener error).
	//
	// Unless the resource has since said otherwise: kubernetes keeps events for
	// about an hour, so an event older than the object's latest condition
	// transition describes a problem that is already over. Honouring it anyway
	// pins a recovered resource to degraded until the event expires.
	if warn != nil && !supersededByStatus(u, warn) {
		if health == healthHealthy || health == healthProgressing {
			health = healthDegraded
		}
		message = warn.reason + ": " + warn.message
	} else {
		warn = nil
	}

	return &models.ServiceComponentHealthResource{
		Provider:     providerKubernetes,
		APIGroup:     gvr.Group,
		Kind:         u.GetKind(),
		Namespace:    u.GetNamespace(),
		Name:         u.GetName(),
		Health:       health,
		Message:      message,
		NativeStatus: native,
		Details:      resourceDetails(u, resourceDiagnosis(u, health, warn)),
	}
}

// supersededByStatus reports whether the object transitioned after the warning
// fired, which means its own status is the newer evidence.
func supersededByStatus(u *unstructured.Unstructured, warn *warningEvent) bool {
	if warn == nil || warn.at.IsZero() {
		return false
	}

	conds, ok, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil || !ok {
		return false
	}
	for _, c := range conds {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		// Only a condition that now reads healthy can retire a warning.
		if status, _ := cond["status"].(string); status != "True" {
			continue
		}
		if t, _ := cond["type"].(string); !isReadyConditionType(t) {
			continue
		}
		raw, _ := cond["lastTransitionTime"].(string)
		at, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		if at.After(warn.at) {
			return true
		}
	}
	return false
}

// componentFor attributes a live object to an install component: manifests by the
// nuon.co ownership labels stamped at apply, helm by release name.
func (e *Engine) componentFor(installID string, u *unstructured.Unstructured) (string, bool) {
	lbls := u.GetLabels()

	if lbls[labelInstallID] == installID {
		if componentID := lbls[labelComponentID]; componentID != "" {
			if _, ok := e.idx.lookup(componentID); ok {
				return componentID, true
			}
		}
	}

	// A terraform module can apply an object directly (kubectl_manifest); it
	// carries no nuon labels and no helm annotations, so nothing else claims it.
	if e.manifestKinds != nil {
		key := resourceKey(u.GetKind(), u.GetNamespace(), u.GetName())
		if componentID, ok := e.manifestKinds.ComponentForObject(key); ok {
			if _, known := e.idx.lookup(componentID); known {
				return componentID, true
			}
		}
	}

	if lbls[helmManagedByLabel] == helmManagedByValue {
		release := u.GetAnnotations()[helmReleaseNameAnnotation]
		if release != "" {
			if entry, ok := e.idx.lookupHelm(release); ok {
				return entry.componentID, true
			}
			// A terraform module can install a chart too; its workloads carry no
			// nuon labels and match no chart component, so they would otherwise
			// be dropped as unowned.
			if e.terraform != nil {
				if componentID, ok := e.terraform.ComponentForRelease(release); ok {
					if _, known := e.idx.lookup(componentID); known {
						return componentID, true
					}
				}
			}
		}
	}

	return "", false
}

// sandboxReleaseFor returns the helm release owning a resource, proven by
// membership in the sandbox's reported release set rather than inferred by
// exclusion, so customer workloads are never surfaced.
func (e *Engine) sandboxReleaseFor(u *unstructured.Unstructured) (string, bool) {
	if u.GetLabels()[helmManagedByLabel] != helmManagedByValue {
		return "", false
	}
	release := u.GetAnnotations()[helmReleaseNameAnnotation]
	if release == "" {
		return "", false
	}
	if !e.cluster.IsSandboxRelease(release) {
		return "", false
	}
	return release, true
}

// clusterWatchedTypes are the component types whose resources this engine
// inspects in the cluster; probes and pushed checks are supplementary for them.
var clusterWatchedTypes = []string{"helm_chart", "kubernetes_manifest"}
