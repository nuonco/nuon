package kubernetes_manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/pkg/diff"
	"github.com/nuonco/nuon/pkg/plans"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
	plantypes "github.com/nuonco/nuon/pkg/types/approvals"
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

	l = l.With(
		zap.String("service.name", "runner.kubernetes_manifest"),
		zap.String("nuon.tool", "kubernetes_manifest"),
		zap.String("nuon.deploy.kind", "kubernetes_manifest"),
		zap.String("k8s.operation", string(job.Operation)),
	)
	ctx = pkgctx.SetLogger(ctx, l)

	if h.clusterProvider != nil {
		h.clusterProvider.Set(h.state.plan.KubernetesManifestDeployPlan.ClusterInfo)
	}

	// why: Hand the applied kinds over too: a manifest of nothing but custom
	// resources is invisible to health otherwise, since nothing knows to list
	// them.
	if h.manifestKinds != nil {
		h.manifestKinds.Set(h.state.plan.ComponentID, h.state.plan.KubernetesManifestDeployPlan.Manifest)
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
		manifest := h.state.plan.KubernetesManifestDeployPlan.Manifest
		return h.handleApplyPlan(ctx, l, k, job, jobExecution, manifest)

	default:
		l.Error("Unsupported operation type", zap.String("operation", string(job.Operation)))
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}
}

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

		liveObj, err := resourceClient.Get(ctx, resource.name, metav1.GetOptions{})

		if err != nil {
			l.Debug("Resource doesn't exist in cluster",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name),
				zap.String("namespace", resource.namespace),
				zap.Error(err))
			continue
		}

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

func (h *handler) resourceDiffWithLive(l *zap.Logger, desired []*kubernetesResource, live []*kubernetesResource) []*kubernetesResource {
	liveMap := make(map[string]*kubernetesResource)
	for _, res := range live {
		key := fmt.Sprintf("%s/%s/%s", res.namespace, res.name, res.groupVersionKind.Kind)
		liveMap[key] = res
	}

	var additions []*kubernetesResource

	ignoreFields := []string{
		"metadata.creationTimestamp",
		"metadata.resourceVersion",
		"metadata.generation",
		"metadata.namespace",
		"metadata.uid",
		"metadata.managedFields",
		"status",
	}

	for _, res := range desired {
		key := fmt.Sprintf("%s/%s/%s", res.namespace, res.name, res.groupVersionKind.Kind)
		liveRes, exists := liveMap[key]

		if !exists {
			additions = append(additions, res)
		} else {
			if liveRes.obj != nil && res.obj != nil {
				changeEntries, hasChanges := diff.DetectChanges(liveRes.obj.Object, res.obj.Object, ignoreFields)

				if hasChanges {
					updatedRes := *res

					pathsChanged := make([]string, 0, len(changeEntries))
					for _, entry := range changeEntries {
						pathsChanged = append(pathsChanged, entry.Path)
					}

					l.Info("resource has changes",
						zap.String("kind", res.groupVersionKind.Kind),
						zap.String("name", res.name),
						zap.String("namespace", res.namespace),
						zap.Strings("paths_changed", pathsChanged),
					)

					additions = append(additions, &updatedRes)
				}
			} else {
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

	liveResources, err := h.fetchLiveResources(ctx, k.client, desiredResources)
	if err != nil {
		l.Error("Failed to fetch live resources from cluster", zap.Error(err))
		return fmt.Errorf("failed to fetch live resources from cluster: %w", err)
	}

	resourcesToApply := h.resourceDiffWithLive(l, desiredResources, liveResources)
	l.Debug("Resource diff calculated against live cluster resources",
		zap.Int("additions/updates", len(resourcesToApply)))

	for i, res := range resourcesToApply {
		l.Debug(fmt.Sprintf("Resource %d to be applied", i),
			zap.String("kind", res.groupVersionKind.Kind),
			zap.String("name", res.name),
			zap.String("namespace", res.namespace))
	}

	l.Info("Performing dry run apply for resources to add/update")

	var resourceDiffs []diff.ResourceDiff

	opCtx, end := op.Tool(ctx, "kubernetes_manifest", "apply_dry_run")
	dryRunApplyOutput, err := h.execApply(opCtx, k.client, resourcesToApply, true)
	end(err)
	if err != nil {
		l.Error("Kubernetes manifest dry run apply failed", zap.Error(err))
		return fmt.Errorf("kubernetes manifest dry run apply failed: %w", err)
	}

	formattedDiffs := diff.FormatResourceDiffs(*dryRunApplyOutput)

	resourceDiffs = append(resourceDiffs, formattedDiffs...)

	manifestPlan.ContentDiff = resourceDiffs

	dryRunYAML, err := kubernetesResourcesToMultiDocYAML(resourcesToApply)
	if err != nil {
		return fmt.Errorf("failed to generate dry run YAML output: %w", err)
	}
	manifestPlan.DryRunOutput = dryRunYAML

	jsonBytes, err := json.MarshalIndent(map[string]interface{}{
		"k8s_content_diff": resourceDiffs,
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

	l.Debug("Marshalled Kubernetes plan contents to JSON",
		zap.Int("diff_entries", len(manifestPlan.ContentDiff)),
		zap.String("operation", string(manifestPlan.Op)))

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
	opCtx, end := op.Tool(ctx, "kubernetes_manifest", "delete_dry_run")
	dryRunDeleteOutput, err := h.execDelete(opCtx, k.client, currentResources, true)
	end(err)
	if err != nil {
		l.Error("Kubernetes manifest dry run delete failed", zap.Error(err))
		return fmt.Errorf("kubernetes manifest dry run delete failed: %w", err)
	}

	formattedDiffs := diff.FormatResourceDiffs(*dryRunDeleteOutput)

	manifestPlan.ContentDiff = formattedDiffs

	dryRunYAML, err := kubernetesResourcesToMultiDocYAML(currentResources)
	if err != nil {
		return fmt.Errorf("failed to generate dry run delete YAML output: %w", err)
	}
	manifestPlan.DryRunOutput = dryRunYAML

	jsonBytes, err := json.MarshalIndent(map[string]interface{}{
		"diff": formattedDiffs,
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

	desiredKubernetesResources, err := h.getKubernetesResourcesFromManifest(k, manifest)
	if err != nil {
		return fmt.Errorf("unable to build kubernetes resources from manifest: %w", err)
	}

	h.state.outputs = map[string]interface{}{"diff": []diff.ResourceDiff{}}

	if manifestPlan.Op == plantypes.KubernetesManifestPlanOperationDelete {
		l.Info("Executing delete operation based on plan",
			zap.Int("resourceCount", len(desiredKubernetesResources)))

		opCtx, end := op.Tool(ctx, "kubernetes_manifest", "delete")
		deleteOutput, err := h.execDelete(opCtx, k.client, desiredKubernetesResources, false)
		end(err)
		if err != nil {
			h.writeErrorResult(ctx, err)
			l.Error("Failed to delete resources", zap.Error(err))
			return fmt.Errorf("failed to delete resources: %w", err)
		}

		formattedDiffs := diff.FormatResourceDiffs(*deleteOutput)

		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]diff.ResourceDiff), formattedDiffs...)

		l.Info("Successfully deleted resources",
			zap.Int("resourceCount", len(desiredKubernetesResources)),
			zap.Int("outputDiffCount", len(*deleteOutput)))
	} else {
		l.Info("Applying resources directly from manifest",
			zap.Int("resourceCount", len(desiredKubernetesResources)))

		opCtx, end := op.Tool(ctx, "kubernetes_manifest", "apply")
		applyOutput, err := h.execApply(opCtx, k.client, desiredKubernetesResources, false)
		end(err)
		if err != nil {
			h.writeErrorResult(ctx, err)
			l.Error("Failed to apply resources", zap.Error(err))
			return fmt.Errorf("failed to apply resources: %w", err)
		}

		formattedDiffs := diff.FormatResourceDiffs(*applyOutput)

		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]diff.ResourceDiff), formattedDiffs...)

		l.Info("Successfully applied resources",
			zap.Int("resourceCount", len(desiredKubernetesResources)),
			zap.Int("outputDiffCount", len(*applyOutput)))
	}

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
	parentL, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logger: %w", err)
	}
	l := parentL

	output := make([]diff.ResourceDiff, 0, len(resources))
	fieldManager := "kube-apply"

	ignoreFields := []string{
		"metadata.creationTimestamp",
		"metadata.resourceVersion",
		"metadata.generation",
		"metadata.namespace",
		"metadata.uid",
		"metadata.managedFields",
		"status",
	}

	for _, resource := range resources {
		if resource == nil {
			parentL.Warn("Skipping nil resource in execApply")
			continue
		}

		l = parentL.With(
			zap.String("k8s.kind", resource.groupVersionKind.Kind),
			zap.String("k8s.name", resource.name),
			zap.String("k8s.namespace", resource.namespace),
		)

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

		liveObj, err := resourceClient.Get(ctx, resource.name, metav1.GetOptions{})
		var currentObj *unstructured.Unstructured
		resourceExists := true
		if err != nil {
			if k8serrors.IsNotFound(err) {
				l.Debug("Resource doesn't exist in cluster yet",
					zap.String("kind", resource.groupVersionKind.Kind),
					zap.String("name", resource.name))
				resourceExists = false
				currentObj = &unstructured.Unstructured{}
			} else {
				l.Warn("Failed to retrieve resource state from cluster, treating as new resource",
					zap.String("kind", resource.groupVersionKind.Kind),
					zap.String("name", resource.name),
					zap.Error(err))
				resourceExists = false
				currentObj = &unstructured.Unstructured{}
			}
		} else {
			currentObj = liveObj
			l.Debug("Retrieved current resource state from cluster",
				zap.String("kind", resource.groupVersionKind.Kind),
				zap.String("name", resource.name))
		}

		originalObj := resource.obj.DeepCopy()

		// why: Remove the managed fields from the object to apply to avoid conflicts
		if originalObj.Object["metadata"] != nil {
			metadata, ok := originalObj.Object["metadata"].(map[string]interface{})
			if ok {
				delete(metadata, "managedFields")

				delete(metadata, "resourceVersion")
				delete(metadata, "generation")
				delete(metadata, "uid")
				delete(metadata, "creationTimestamp")
			}
		}

		if h.state.plan.InstallID != "" || h.state.plan.ComponentID != "" {
			objLabels := originalObj.GetLabels()
			if objLabels == nil {
				objLabels = map[string]string{}
			}
			if h.state.plan.InstallID != "" {
				objLabels["nuon.co/install-id"] = h.state.plan.InstallID
			}
			if h.state.plan.ComponentID != "" {
				objLabels["nuon.co/component-id"] = h.state.plan.ComponentID
			}
			originalObj.SetLabels(objLabels)
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
			changeEntries, hasChanges := diff.DetectChanges(currentObj.Object, appliedObj.Object, ignoreFields)

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

				op.Entries = []diff.DiffEntry{
					{
						Type:    diff.EntryAdded,
						Applied: originalObj.Object,
					},
				}
			} else {
				op.Type = diff.EntryModified

				changeEntries, _ := diff.DetectChanges(currentObj.Object, appliedObj.Object, ignoreFields)
				op.Entries = changeEntries
			}
		}

		output = append(output, op)
	}

	l.Debug("execApply finished",
		zap.Int("output_entries", len(output)),
		zap.Int("resource_count", len(resources)))
	return &output, nil
}

func (h *handler) execDelete(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource, dryRun bool) (*[]diff.ResourceDiff, error) {
	output := make([]diff.ResourceDiff, 0, len(resources))

	for _, resource := range resources {
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
