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
			byKey[resourceKey(u.GetKind(), u.GetNamespace(), u.GetName())] = u
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
	failed := failedPods(byKey)
	e.hydrateFailedPodOwners(ctx, dynClient, failed, byKey)

	lifted := liftPodHealthToOwners(failed, byKey)

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

		key := resourceKey(u.GetKind(), u.GetNamespace(), u.GetName())
		var warn *warningEvent
		if w, ok := warnings[key]; ok {
			warn = &w
		}
		res := resourceModel(gvr, u, warn, lifted[key])

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

// hydrateFailedPodOwners GETs the owner chain of each pod whose own status
// reports it failing, so byKey can resolve it. Normally nothing is fetched:
// healthy installs have no failing pods, which is the point of doing this on
// demand instead of listing every ReplicaSet in the cluster once a minute.
func (e *Engine) hydrateFailedPodOwners(
	ctx context.Context,
	dynClient dynamic.Interface,
	failed []podFailure,
	byKey map[string]*unstructured.Unstructured,
) {
	for _, f := range failed {
		cur := byKey[f.key]
		for hop := 0; hop < maxOwnerHops; hop++ {
			ref := metav1.GetControllerOf(cur)
			if ref == nil {
				break
			}
			ownerKey := resourceKey(ref.Kind, cur.GetNamespace(), ref.Name)
			if next, ok := byKey[ownerKey]; ok {
				cur = next
				continue
			}
			gvr, ok := ownerGVRs[ref.Kind]
			if !ok {
				break
			}
			owner, err := dynClient.Resource(gvr).
				Namespace(cur.GetNamespace()).
				Get(ctx, ref.Name, metav1.GetOptions{})
			if err != nil {
				e.l.Warn("unable to fetch owner for failed pod",
					zap.String("kind", ref.Kind), zap.String("name", ref.Name), zap.Error(err))
				break
			}
			byKey[ownerKey] = owner
			cur = owner
		}
	}
}

// liftPodHealthToOwners maps a controller to the failure of the first of its
// pods that reports one, so an ImagePullBackOff surfaces immediately instead of
// waiting out the Deployment's progressDeadlineSeconds.
func liftPodHealthToOwners(failed []podFailure, byKey map[string]*unstructured.Unstructured) map[string]string {
	lifted := map[string]string{}
	for _, f := range failed {
		owner := topOwner(byKey[f.key], byKey)
		if owner == nil {
			continue
		}
		ownerKey := resourceKey(owner.GetKind(), owner.GetNamespace(), owner.GetName())
		if _, seen := lifted[ownerKey]; seen {
			continue
		}
		lifted[ownerKey] = f.message
	}
	return lifted
}

type podFailure struct {
	key     string
	message string
}

// failedPods returns the pods whose own status reports them failing, sorted so
// that which pod explains a shared owner is deterministic rather than a product
// of map iteration order.
//
// Only a pod its own status calls degraded qualifies: one still starting up is
// progressing, and lifting that would degrade every rollout the moment it began.
func failedPods(byKey map[string]*unstructured.Unstructured) []podFailure {
	var out []podFailure
	for key, u := range byKey {
		if u.GetKind() != "Pod" || terminalPodPhase(u) {
			continue
		}
		health, message, _ := assessResource(u)
		if health != healthDegraded {
			continue
		}
		if message == "" {
			message = "pod " + u.GetName() + " is not healthy"
		}
		out = append(out, podFailure{key: key, message: message})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out
}

// terminalPodPhase reports a pod that has finished and been replaced.
// Kubernetes keeps failed pods until the GC threshold is hit, so an evicted one
// lingers for hours — lifting it would hold a controller degraded long after
// its replacement came up healthy.
func terminalPodPhase(u *unstructured.Unstructured) bool {
	phase, _, _ := unstructured.NestedString(u.Object, "status", "phase")
	return phase == "Failed" || phase == "Succeeded"
}

// topOwner returns the furthest controller ancestor present in the listed
// objects, or nil when the object has no owner in the set.
func topOwner(u *unstructured.Unstructured, byKey map[string]*unstructured.Unstructured) *unstructured.Unstructured {
	cur := u
	for hop := 0; hop < maxOwnerHops; hop++ {
		ref := metav1.GetControllerOf(cur)
		if ref == nil {
			break
		}
		next, ok := byKey[resourceKey(ref.Kind, cur.GetNamespace(), ref.Name)]
		if !ok || next == cur {
			break
		}
		cur = next
	}
	if cur == u {
		return nil
	}
	return cur
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

func resourceModel(
	gvr schema.GroupVersionResource,
	u *unstructured.Unstructured,
	warn *warningEvent,
	liftedFailure string,
) *models.ServiceComponentHealthResource {
	health, message, native := assessResource(u)

	// A pod's failure is its controller's failure: helm annotates only what it
	// renders, so pods carry no ownership labels and are never attributed to a
	// component on their own.
	if liftedFailure != "" && (health == healthHealthy || health == healthProgressing) {
		health = healthDegraded
		message = liftedFailure
	}

	// An event never decides the verdict. Events are edge-triggered and have no
	// "all clear", so consulting one forces us to invent an expiry, and any
	// expiry is a fabricated claim about how long the fault lasted — a single
	// transient event pinned components degraded for exactly 15m + 2 reports
	// however fast the resource recovered. An event only explains a resource
	// that its own status already reports as failing.
	if health == healthHealthy {
		warn = nil
	} else if warn != nil && message == "" {
		message = warn.reason + ": " + warn.message
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
