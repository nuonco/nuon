package kubernetes_manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/diff"
	"github.com/powertoolsdev/mono/pkg/plans"
	plantypes "github.com/powertoolsdev/mono/pkg/types/components/plan"
)

type kubernetesResource struct {
	groupVersionKind     schema.GroupVersionKind
	groupVersionResource schema.GroupVersionResource
	namespace            string
	name                 string
	raw                  string
	obj                  *unstructured.Unstructured
	namespaced           bool
}

func (h *handler) getApplyPlanContents(l *zap.Logger, applyPlanContents string) ([]byte, error) {
	l.Info("decoding and decompressing apply plan contents", zap.Int("contents.string.length", len(applyPlanContents)))
	decompressedBytes, err := plans.DecompressPlan(applyPlanContents)
	if err != nil {
		return []byte{}, errors.Wrap(err, "unable to decode or decompress apply plan contents")
	}

	l.Info("decompressed apply plan contents", zap.Int("contents.decompressed_bytes.length", len(decompressedBytes)))
	return decompressedBytes, nil
}

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Debug("Starting Exec function",
		zap.String("jobID", job.ID),
		zap.String("operation", string(job.Operation)))

	k := h.state.kubeClient

	desiredKubernetesResources, err := h.getKubernetesResourcesFromManifest(
		k,
		h.state.plan.KubernetesManifestDeployPlan.Manifest,
	)
	if err != nil {
		return fmt.Errorf("unable to build kubernetes resources from raw manifest: %w", err)
	}
	l.Debug("Desired Kubernetes resources from manifest",
		zap.Int("resourceCount", len(desiredKubernetesResources)))

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		return h.handleCreateApplyPlan(ctx, l, k, desiredKubernetesResources)

	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		return h.handleCreateTeardownPlan(ctx, l, k, desiredKubernetesResources)

	case models.AppRunnerJobOperationTypeApplyDashPlan:
		// Get manifest from the plan
		manifest := h.state.plan.KubernetesManifestDeployPlan.Manifest
		return h.handleApplyPlan(ctx, l, k, job, jobExecution, manifest)

	default:
		l.Error("Unsupported operation type", zap.String("operation", string(job.Operation)))
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}
}

// fetchLiveResources gets the current state of resources from the cluster
func (h *handler) fetchLiveResources(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource) ([]*kubernetesResource, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	liveResources := make([]*kubernetesResource, 0, len(resources))

	for _, resource := range resources {
		var resourceClient dynamic.ResourceInterface
		if resource.namespaced {
			resourceClient = client.Resource(resource.groupVersionResource).Namespace(resource.namespace)
		} else {
			resourceClient = client.Resource(resource.groupVersionResource)
		}

		// Try to get the resource from the cluster
		liveObj, err := resourceClient.Get(ctx, resource.name, metav1.GetOptions{})

		if err != nil {
			// Resource doesn't exist in the cluster
			l.Debug("Resource doesn't exist in cluster",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name),
				zap.String("namespace", resource.namespace),
				zap.Error(err))
			continue
		}

		// Resource exists, add it to our list
		liveResource := &kubernetesResource{
			groupVersionKind:     resource.groupVersionKind,
			groupVersionResource: resource.groupVersionResource,
			namespace:            resource.namespace,
			name:                 resource.name,
			obj:                  liveObj,
			namespaced:           resource.namespaced,
		}

		if liveObj != nil && liveObj.Object != nil {
			objBytes, err := json.Marshal(liveObj.Object)
			if err == nil {
				liveResource.raw = string(objBytes)
			}
		}

		liveResources = append(liveResources, liveResource)
		l.Debug("Found live resource in cluster",
			zap.String("kind", resource.groupVersionKind.Kind),
			zap.String("name", resource.name),
			zap.String("namespace", resource.namespace))
	}

	l.Info("Fetched live resources from cluster",
		zap.Int("desiredCount", len(resources)),
		zap.Int("liveCount", len(liveResources)))
	return liveResources, nil
}

// resourceDiffWithLive compares desired resources with live resources to determine what needs to be added/updated
func (h *handler) resourceDiffWithLive(desired []*kubernetesResource, live []*kubernetesResource) []*kubernetesResource {
	// Map of live resources by name+namespace+kind for quick lookup
	liveMap := make(map[string]*kubernetesResource)
	for _, res := range live {
		key := fmt.Sprintf("%s/%s/%s", res.namespace, res.name, res.groupVersionKind.Kind)
		liveMap[key] = res
	}

	// Resources to add or update
	var additions []*kubernetesResource

	// Fields to ignore when comparing resources
	ignoreFields := []string{
		"metadata.creationTimestamp",
		"metadata.resourceVersion",
		"metadata.namespace",
		"metadata.uid",
		"metadata.managedFields",
		"status",
	}

	// Find resources to add or update
	for _, res := range desired {
		key := fmt.Sprintf("%s/%s/%s", res.namespace, res.name, res.groupVersionKind.Kind)
		liveRes, exists := liveMap[key]

		if !exists {
			// Resource doesn't exist in the cluster - needs to be added
			additions = append(additions, res)
		} else {
			// Resource exists - check if it needs updating by comparing the objects
			if liveRes.obj != nil && res.obj != nil {
				// Use DetectChanges to get detailed change information
				changeEntries, hasChanges := diff.DetectChanges(liveRes.obj.Object, res.obj.Object, ignoreFields)

				if hasChanges {
					// Store the change entries in the resource for later use
					// We'll create a copy of the resource to avoid modifying the original
					updatedRes := *res // Create a copy

					// We could store the change entries in the resource itself
					// This would require adding a field to kubernetesResource
					// Or we can just add the resource to our list for now

					// Log some details about what's changing
					pathsChanged := make([]string, 0, len(changeEntries))
					for _, entry := range changeEntries {
						pathsChanged = append(pathsChanged, entry.Path)
					}

					fmt.Printf("Resource %s/%s of kind %s has changes in paths: %v\n",
						res.namespace, res.name, res.groupVersionKind.Kind, pathsChanged)

					additions = append(additions, &updatedRes)
				}
				// If no changes, we don't need to do anything with this resource
			} else {
				// If we can't properly compare the objects, add it to be safe
				additions = append(additions, res)
			}
		}
	}

	return additions
}

func (h *handler) handleCreateApplyPlan(
	ctx context.Context,
	l *zap.Logger,
	k *kubernetesClient,
	desiredResources []*kubernetesResource,
) error {
	var manifestPlan plantypes.KubernetesManifestPlanContents

	l.Debug("Processing Create-Apply-Plan operation")
	manifestPlan.Op = plantypes.KubernetesManifestPlanOperationApply

	// Fetch live resources from the cluster for comparison
	liveResources, err := h.fetchLiveResources(ctx, k.client, desiredResources)
	if err != nil {
		l.Error("Failed to fetch live resources from cluster", zap.Error(err))
		return fmt.Errorf("failed to fetch live resources from cluster: %w", err)
	}

	// Determine resources to add/update based on comparison with live resources
	// note: we do not track deletions in this. Deletions are handled only in teardown plans.
	resourcesToApply := h.resourceDiffWithLive(desiredResources, liveResources)
	l.Debug("Resource diff calculated against live cluster resources",
		zap.Int("additions/updates", len(resourcesToApply)))

	// Log details about which resources will be updated
	for i, res := range resourcesToApply {
		l.Debug(fmt.Sprintf("Resource %d to be applied", i),
			zap.String("kind", res.groupVersionKind.Kind),
			zap.String("name", res.name),
			zap.String("namespace", res.namespace))
	}

	// Run dry run apply
	l.Info("Performing dry run apply for resources to add/update")

	// This will contain our detailed change information
	var resourceDiffs []diff.ResourceDiff

	// Execute dry run apply to get detailed diffs
	dryRunApplyOutput, err := h.execApply(ctx, k.client, resourcesToApply, true)
	if err != nil {
		l.Error("Kubernetes manifest dry run apply failed", zap.Error(err))
		return fmt.Errorf("kubernetes manifest dry run apply failed: %w", err)
	}

	// Add all apply diffs to our resource diffs
	resourceDiffs = append(resourceDiffs, *dryRunApplyOutput...)

	// Store the detailed diffs in the plan
	manifestPlan.ContentDiff = resourceDiffs

	// Convert to JSON for the plan field
	jsonBytes, err := json.MarshalIndent(map[string]interface{}{
		"diff": resourceDiffs,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal combined dry run results to JSON: %w", err)
	}

	manifestPlan.Plan = string(jsonBytes)

	l.Info("Kubernetes manifest dry run completed",
		zap.Int("resource_count", len(desiredResources)),
		zap.Int("apply_diff_entries", len(*dryRunApplyOutput)),
		zap.Int("total_diff_entries", len(resourceDiffs)))

	planJ, err := json.Marshal(manifestPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal k8s plan contents to JSON: %w", err)
	}

	// Encode and submit results
	encodedPlan, err := plans.CompressPlan(planJ)
	if err != nil {
		return fmt.Errorf("failed to compress plan: %w", err)
	}

	apiRes := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:                   true,
		ContentsCompressed:        encodedPlan,
		ContentsDisplayCompressed: encodedPlan,
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, apiRes)
	if err != nil {
		l.Error("Failed to create job execution result", zap.Error(err))
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) handleCreateTeardownPlan(
	ctx context.Context,
	l *zap.Logger,
	k *kubernetesClient,
	currentResources []*kubernetesResource,
) error {
	l.Debug("Processing Create-Teardown-Plan operation")
	var manifestPlan plantypes.KubernetesManifestPlanContents

	manifestPlan.Op = plantypes.KubernetesManifestPlanOperationDelete

	l.Info("Performing dry run delete for teardown plan")
	dryRunDeleteOutput, err := h.execDelete(ctx, k.client, currentResources, true)
	if err != nil {
		l.Error("Kubernetes manifest dry run delete failed", zap.Error(err))
		return fmt.Errorf("kubernetes manifest dry run delete failed: %w", err)
	}

	manifestPlan.ContentDiff = *dryRunDeleteOutput

	jsonBytes, err := json.MarshalIndent(map[string]interface{}{
		"diff": *dryRunDeleteOutput,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dry run delete results to JSON: %w", err)
	}

	manifestPlan.Plan = string(jsonBytes)
	l.Info("Kubernetes manifest dry run delete succeeded",
		zap.Int("resource_count", len(currentResources)),
		zap.Int("diff_entries", len(*dryRunDeleteOutput)))

	planJ, err := json.Marshal(manifestPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal k8s plan contents to JSON: %w", err)
	}

	// Encode and submit results
	encodedPlan, err := plans.CompressPlan(planJ)
	if err != nil {
		return fmt.Errorf("failed to compress plan: %w", err)
	}

	apiRes := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:                   true,
		ContentsCompressed:        encodedPlan,
		ContentsDisplayCompressed: encodedPlan,
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, apiRes)
	if err != nil {
		l.Error("Failed to create job execution result", zap.Error(err))
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) handleApplyPlan(
	ctx context.Context,
	l *zap.Logger,
	k *kubernetesClient,
	job *models.AppRunnerJob,
	jobExecution *models.AppRunnerJobExecution,
	manifest string,
) error {
	l.Debug("Processing Apply-Plan operation with manifest directly")

	var manifestPlan plantypes.KubernetesManifestPlanContents

	// Decode the plan contents to get the operation type
	planContents, err := h.getApplyPlanContents(l, h.state.plan.ApplyPlanContents)
	if err != nil {
		return errors.Wrap(err, "unable to get apply plan contents")
	}

	if err := json.Unmarshal(planContents, &manifestPlan); err != nil {
		return errors.Wrap(err, "unable to decode apply plan")
	}

	l.Debug("Apply plan decoded",
		zap.String("operation", string(manifestPlan.Op)),
		zap.Int("content_diff_count", len(manifestPlan.ContentDiff)))

	// Get Kubernetes resources from the manifest
	desiredKubernetesResources, err := h.getKubernetesResourcesFromManifest(k, manifest)
	if err != nil {
		return fmt.Errorf("unable to build kubernetes resources from manifest: %w", err)
	}

	// Initialize outputs
	h.state.outputs = map[string]interface{}{"diff": []diff.ResourceDiff{}}

	// Check if this is a delete operation
	if manifestPlan.Op == plantypes.KubernetesManifestPlanOperationDelete {
		l.Info("Executing delete operation based on plan",
			zap.Int("resourceCount", len(desiredKubernetesResources)))

		// Execute delete operation for all resources
		deleteOutput, err := h.execDelete(ctx, k.client, desiredKubernetesResources, false)
		if err != nil {
			h.writeErrorResult(ctx, err)
			l.Error("Failed to delete resources", zap.Error(err))
			return fmt.Errorf("failed to delete resources: %w", err)
		}

		// Store the results
		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]diff.ResourceDiff), *deleteOutput...)

		l.Info("Successfully deleted resources",
			zap.Int("resourceCount", len(desiredKubernetesResources)),
			zap.Int("outputDiffCount", len(*deleteOutput)))
	} else {
		// Default case: Apply operation (existing logic)
		l.Info("Applying resources directly from manifest",
			zap.Int("resourceCount", len(desiredKubernetesResources)))

		// Execute apply operation for all resources
		applyOutput, err := h.execApply(ctx, k.client, desiredKubernetesResources, false)
		if err != nil {
			h.writeErrorResult(ctx, err)
			l.Error("Failed to apply resources", zap.Error(err))
			return fmt.Errorf("failed to apply resources: %w", err)
		}

		// Store the results
		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]diff.ResourceDiff), *applyOutput...)

		l.Info("Successfully applied resources",
			zap.Int("resourceCount", len(desiredKubernetesResources)),
			zap.Int("outputDiffCount", len(*applyOutput)))
	}

	// Create API result
	apiRes, err := h.createAPIResultRequest(nil, l, manifestPlan)
	if err != nil {
		h.writeErrorResult(ctx, err)
		l.Error("Failed to create API result", zap.Error(err))
		return fmt.Errorf("unable to create api result: %w", err)
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes)
	if err != nil {
		l.Error("Failed to create job execution result", zap.Error(err))
		h.errRecorder.Record("write job execution result", err)
	}

	return nil
}

func (h *handler) execApply(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource, dryRun bool) (*[]diff.ResourceDiff, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}

	output := make([]diff.ResourceDiff, 0, len(resources))
	fieldManager := "kube-apply"

	ignoreFields := []string{
		"metadata.creationTimestamp",
		"metadata.resourceVersion",
		"metadata.namespace",
		"metadata.uid",
		"metadata.managedFields",
		"status",
	}

	for _, resource := range resources {
		if resource == nil {
			l.Warn("Skipping nil resource in execApply")
			continue
		}

		if resource.obj == nil {
			l.Error("Resource object is nil, cannot apply",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name),
				zap.String("namespace", resource.namespace))

			output = append(output, diff.ResourceDiff{
				Name:      resource.name,
				Namespace: resource.namespace,
				Kind:      resource.groupVersionKind.Kind,
				ApiPath:   fmt.Sprintf("%s/%s", resource.groupVersionResource.Group, resource.groupVersionResource.Version),
				Resource:  resource.groupVersionResource.Resource,
				Operation: string(plantypes.KubernetesManifestPlanOperationApply),
				DryRun:    dryRun,
				Version:   "2",
				Type:      diff.EntryError,
				ErrorMsg:  "Resource object is nil, cannot apply",
			})
			continue
		}

		op := diff.ResourceDiff{
			Name:      resource.name,
			Namespace: resource.namespace,
			Kind:      resource.groupVersionKind.Kind,
			ApiPath:   fmt.Sprintf("%s/%s", resource.groupVersionResource.Group, resource.groupVersionResource.Version),
			Resource:  resource.groupVersionResource.Resource,
			Operation: string(plantypes.KubernetesManifestPlanOperationApply),
			DryRun:    dryRun,
			Version:   "2",
			Type:      diff.EntryModified,
		}

		var resourceClient dynamic.ResourceInterface
		if resource.namespaced {
			resourceClient = client.Resource(resource.groupVersionResource).Namespace(resource.namespace)
		} else {
			resourceClient = client.Resource(resource.groupVersionResource)
		}

		// Get the current live state of the resource from the cluster
		liveObj, err := resourceClient.Get(ctx, resource.name, metav1.GetOptions{})
		var currentObj *unstructured.Unstructured
		resourceExists := true
		if err != nil {
			// Resource might not exist yet
			l.Debug("Resource doesn't exist in cluster yet or can't be accessed",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name),
				zap.Error(err))
			resourceExists = false
			// We'll create a placeholder for comparison
			currentObj = &unstructured.Unstructured{}
		} else {
			currentObj = liveObj
			l.Debug("Retrieved current resource state from cluster",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name))
		}

		originalObj := resource.obj.DeepCopy()

		// Remove the managed fields from the object to apply to avoid conflicts
		if originalObj.Object["metadata"] != nil {
			metadata, ok := originalObj.Object["metadata"].(map[string]interface{})
			if ok {
				// Remove the managedFields to prevent the error
				delete(metadata, "managedFields")

				// Also ensure other server-set fields don't cause issues
				delete(metadata, "resourceVersion")
				delete(metadata, "generation")
				delete(metadata, "uid")
				delete(metadata, "creationTimestamp")
			}
		}

		applyOptions := metav1.ApplyOptions{
			FieldManager: fieldManager,
			Force:        true,
		}
		if dryRun {
			applyOptions.DryRun = []string{"All"}
			l.Debug("Performing dry run apply",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name))
		}

		appliedObj, err := resourceClient.Apply(ctx, resource.name, originalObj, applyOptions)
		if err != nil {
			op.Type = diff.EntryError
			op.ErrorMsg = err.Error()
			output = append(output, op)

			if !dryRun {
				return &output, fmt.Errorf("apply error for resource [%s %s/%s]: %w",
					resource.groupVersionKind.Kind, resource.namespace, resource.name, err)
			}
			continue
		}

		if dryRun {
			// Get detailed diff entries between current state and desired state
			// This is the key change - we use the diff package's DetectChanges function
			changeEntries, hasChanges := diff.DetectChanges(currentObj.Object, originalObj.Object, ignoreFields)

			// Add these detailed diff entries to our ResourceDiff
			op.Entries = changeEntries

			if !resourceExists {
				op.Type = diff.EntryAdded
			} else if hasChanges {
				op.Type = diff.EntryModified
			} else {
				op.Type = diff.EntryUnchanged
			}
		} else {
			if !resourceExists {
				op.Type = diff.EntryAdded

				// For actual applies, still capture what was applied
				op.Entries = []diff.DiffEntry{
					{
						Type:    diff.EntryAdded,
						Applied: originalObj.Object,
					},
				}
			} else {
				op.Type = diff.EntryModified

				// For actual applies of existing resources, capture before/after
				changeEntries, _ := diff.DetectChanges(currentObj.Object, appliedObj.Object, ignoreFields)
				op.Entries = changeEntries
			}
		}

		output = append(output, op)
	}

	return &output, nil
}

func (h *handler) execDelete(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource, dryRun bool) (*[]diff.ResourceDiff, error) {
	output := make([]diff.ResourceDiff, 0, len(resources))

	for _, resource := range resources {
		// Skip invalid resources
		if resource == nil {
			continue
		}

		op := diff.ResourceDiff{
			Name:      resource.name,
			Namespace: resource.namespace,
			Kind:      resource.groupVersionKind.Kind,
			ApiPath:   fmt.Sprintf("%s/%s", resource.groupVersionResource.Group, resource.groupVersionResource.Version),
			Resource:  resource.groupVersionResource.Resource,
			Operation: string(plantypes.KubernetesManifestPlanOperationDelete),
			DryRun:    dryRun,
			Version:   "2",
			Type:      diff.EntryRemoved,
		}

		entry := diff.DiffEntry{
			Type: diff.EntryRemoved,
		}

		if resource.obj != nil {
			entry.Original = resource.obj.Object
		} else {
			entry.Original = map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/%s", resource.groupVersionResource.Group, resource.groupVersionResource.Version),
				"kind":       resource.groupVersionKind.Kind,
				"metadata": map[string]interface{}{
					"name":      resource.name,
					"namespace": resource.namespace,
				},
			}
		}
		op.Entries = append(op.Entries, entry)

		if dryRun {

			output = append(output, op)
			continue
		}

		var resourceClient dynamic.ResourceInterface
		if resource.namespaced {
			resourceClient = client.Resource(resource.groupVersionResource).Namespace(resource.namespace)
		} else {
			resourceClient = client.Resource(resource.groupVersionResource)
		}

		// Create delete options and add DryRun if needed
		deleteOptions := metav1.DeleteOptions{}
		if dryRun {
			deleteOptions.DryRun = []string{"All"}
		}

		err := resourceClient.Delete(ctx, resource.name, deleteOptions)
		if err != nil {
			op.Type = diff.EntryError
			op.ErrorMsg = err.Error()
			output = append(output, op)

			return &output, fmt.Errorf("delete error for resource [%s %s/%s]: %w",
				resource.groupVersionKind.Kind, resource.namespace, resource.name, err)
		}

		output = append(output, op)
	}

	return &output, nil
}
