package componenthealth

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const providerKubernetes = "kubernetes"

// report runs one stateless cycle: refresh the index, build a fresh cluster
// client, list watched resources + warning events, assess each, and push.
func (e *Engine) report(ctx context.Context) error {
	if err := e.rebuildIndex(ctx); err != nil {
		return fmt.Errorf("unable to refresh component index: %w", err)
	}
	installID := e.idx.installIDValue()
	if installID == "" {
		return nil
	}

	ci := e.cluster.Get()
	if ci == nil {
		e.l.Info("component health: waiting for cluster access (not yet provided by a deploy)")
		return nil
	}
	restCfg, err := kube.ConfigForCluster(ctx, ci)
	if err != nil {
		return fmt.Errorf("unable to build cluster config: %w", err)
	}
	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("unable to build dynamic client: %w", err)
	}

	warnings := e.latestWarnings(ctx, dynClient)

	grouped := map[string][]*models.ServiceComponentHealthResource{}
	sandbox := map[string][]*models.ServiceComponentHealthResource{}
	sandboxNS := map[string]string{}

	for _, gvr := range watchedGVRs {
		list, err := dynClient.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			e.l.Warn("unable to list resources for component health",
				zap.String("resource", gvr.String()), zap.Error(err))
			continue
		}
		for i := range list.Items {
			u := &list.Items[i]

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
			if w, ok := warnings[resourceKey(u.GetKind(), u.GetNamespace(), u.GetName())]; ok {
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
	}

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
		components = append(components, &models.ServiceComponentHealthComponent{
			InstallComponentID: &installComponentID,
			ComponentID:        entry.componentID,
			ComponentType:      entry.componentType,
			Resources:          resources,
		})
	}

	sandboxReleases := make([]*models.ServiceComponentHealthSandboxRelease, 0, len(sandbox))
	for release, resources := range sandbox {
		releaseName := release
		sandboxReleases = append(sandboxReleases, &models.ServiceComponentHealthSandboxRelease{
			ReleaseName: &releaseName,
			Namespace:   sandboxNS[release],
			Resources:   resources,
		})
	}

	req := &models.ServiceCreateComponentHealthRequest{
		InstallID:       &installID,
		Kind:            "resync",
		ObservedAt:      time.Now().UTC().Format(time.RFC3339),
		Components:      components,
		SandboxReleases: sandboxReleases,
	}
	if _, err := e.apiClient.CreateComponentHealth(ctx, req); err != nil {
		return fmt.Errorf("unable to report component health: %w", err)
	}

	return nil
}

func resourceModel(gvr schema.GroupVersionResource, u *unstructured.Unstructured, warn *warningEvent) *models.ServiceComponentHealthResource {
	health, message, native := assessResource(u)

	// A current Warning event means the resource is failing even when its own
	// status looks benign (e.g. an Ingress stuck Progressing on a listener
	// error). Reflect it in the verdict and surface the event as the message.
	if warn != nil {
		if health == healthHealthy || health == healthProgressing {
			health = healthDegraded
		}
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

// componentFor attributes a live object to an install component: kube-manifest
// resources via the nuon.co ownership labels stamped at apply, helm resources by
// release name.
func (e *Engine) componentFor(installID string, u *unstructured.Unstructured) (string, bool) {
	lbls := u.GetLabels()

	if lbls[labelInstallID] == installID {
		if componentID := lbls[labelComponentID]; componentID != "" {
			if _, ok := e.idx.lookup(componentID); ok {
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
		}
	}

	return "", false
}

// sandboxReleaseFor returns the helm release name for a resource the install's
// sandbox owns (base infra), proven by membership in the release set the sandbox
// apply reported — never inferred by exclusion, so customer workloads are never
// surfaced.
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
