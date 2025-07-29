package kubernetes_manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/generics"
	types "github.com/powertoolsdev/mono/pkg/types/components/plan"
	"go.uber.org/zap"
	release "helm.sh/helm/v4/pkg/release/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
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

// NOTE: the helm plans are not real plans, they are just diffs
func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Debug("Starting Exec function",
		zap.String("jobID", job.ID),
		zap.String("operation", string(job.Operation)),
	)

	var kubernetesManifestPlan types.KubernetesManifestPlanContents
	var rel *release.Release

	k := h.state.kubeClient

	l.Debug("Kubernetes client initialized", zap.String("jobID", job.ID))

	// Get current Kubernetes resources
	currentKubernetesResources, err := h.getKubernetesResourcesFromManifest(
		ctx,
		k,
		h.state.plan.KubernetesManifestDeployPlan.Manifest,
	)
	if err != nil {
		return fmt.Errorf("unable to build kubernetes resources from raw manifest: %w", err)
	}

	l.Debug("Current Kubernetes resources fetched",
		zap.Int("resourceCount", len(currentKubernetesResources)),
		zap.String("jobID", job.ID),
	)

	// Get previous Kubernetes resources
	var previousConfigKubernetesResources []*kubernetesResource
	if h.state.previousDeployResources != nil {
		previousConfigKubernetesResources, err = h.getKubernetesResourcesFromManifest(
			ctx,
			k,
			generics.FromPtrStr(h.state.previousDeployResources),
		)
		if err != nil {
			return fmt.Errorf("unable to build previous config kubernetes resources from raw manifest: %w", err)
		}

		l.Debug("Previous Kubernetes resources fetched",
			zap.Int("resourceCount", len(previousConfigKubernetesResources)),
			zap.String("jobID", job.ID),
		)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		l.Debug("Processing Create-Apply-Plan operation", zap.String("jobID", job.ID))

		addition, deletion := h.resourceDiff(
			previousConfigKubernetesResources,
			currentKubernetesResources,
		)

		l.Debug("Resource diff calculated",
			zap.Int("additions", len(addition)),
			zap.Int("deletions", len(deletion)),
			zap.String("jobID", job.ID),
		)

		for _, r := range addition {
			plan := types.KubernetesManifestDiff{}
			plan.Op = types.KubernetesManifestPlanOperationApply
			plan.GroupVersionKind = r.groupVersionKind
			plan.GroupVersionResource = r.groupVersionResource
			plan.Namespace = r.namespace
			plan.Name = r.name
			plan.Before = h.getPreviousConfigForResource(ctx, &r, previousConfigKubernetesResources).raw
			plan.After = r.raw

			kubernetesManifestPlan.Plan = append(kubernetesManifestPlan.Plan, &plan)

			h.state.outputs = nil
		}

		for _, r := range deletion {
			plan := types.KubernetesManifestDiff{}
			plan.Op = types.KubernetesManifestPlanOperationDelete
			plan.GroupVersionKind = r.groupVersionKind
			plan.GroupVersionResource = r.groupVersionResource
			plan.Namespace = r.namespace
			plan.Name = r.name
			plan.Before = h.getPreviousConfigForResource(ctx, &r, previousConfigKubernetesResources).raw

			kubernetesManifestPlan.Plan = append(kubernetesManifestPlan.Plan, &plan)
			h.state.outputs = nil
		}

	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Debug("Processing Create-Teardown-Plan operation", zap.String("jobID", job.ID))

		for _, r := range currentKubernetesResources {
			plan := types.KubernetesManifestDiff{}
			plan.Op = types.KubernetesManifestPlanOperationDelete
			plan.GroupVersionKind = r.groupVersionKind
			plan.GroupVersionResource = r.groupVersionResource
			plan.Namespace = r.namespace
			plan.Name = r.name
			plan.Before = r.raw

			kubernetesManifestPlan.Plan = append(kubernetesManifestPlan.Plan, &plan)
		}

	case models.AppRunnerJobOperationTypeApplyDashPlan:
		l.Debug("Processing Apply-Plan operation", zap.String("jobID", job.ID))

		applyPlan := types.KubernetesManifestPlanContents{}
		err := json.Unmarshal([]byte(h.state.plan.ApplyPlanContents), &applyPlan)
		if err != nil {
			return fmt.Errorf("unable to decode apply plan %w", err)
		}

		l.Debug("Apply plan decoded",
			zap.Int("planCount", len(applyPlan.Plan)),
			zap.String("jobID", job.ID),
		)

		var ar []*kubernetesResource
		var dr []*kubernetesResource

		for _, diff := range applyPlan.Plan {
			switch diff.Op {
			case types.KubernetesManifestPlanOperationApply:
				r, err := h.getKubernetesResourcesFromManifest(ctx, k, diff.After)
				if err != nil {
					err = errors.Wrap(err, "unable to build kubernetes resource from manifest %w")
					break
				}
				ar = append(ar, r...)
			case types.KubernetesManifestPlanOperationDelete:
				r, err := h.getKubernetesResourcesFromManifest(ctx, k, diff.Before)
				if err != nil {
					err = errors.Wrap(err, "unable to build kubernetes resource from manifest %w")
					break
				}
				dr = append(dr, r...)
			default:
				// case not possible
			}
		}

		l.Debug("Resources grouped for apply and delete",
			zap.Int("applyCount", len(ar)),
			zap.Int("deleteCount", len(dr)),
			zap.String("jobID", job.ID),
		)

		h.state.outputs = map[string]interface{}{"diff": []operationOutput{}}
		output, err := h.execApply(ctx, k.client, ar)
		if err != nil {
			return err
		}
		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]operationOutput), *output...)
		output, err = h.execDelete(ctx, k.client, dr)
		if err != nil {
			return err
		}
		h.state.outputs["diff"] = append(h.state.outputs["diff"].([]operationOutput), *output...)

		kubernetesManifestPlan = applyPlan

	default:
		l.Error("Unsupported operation type",
			zap.String("operation", string(job.Operation)),
			zap.String("jobID", job.ID),
		)
		return fmt.Errorf("unsupported run type %s", job.Operation)
	}

	// Handle error
	if err != nil {
		h.writeErrorResult(ctx, err)
		l.Error("Error occurred during execution", zap.Error(err), zap.String("jobID", job.ID))
		return fmt.Errorf("unable to run plan, %w", err)
	}

	// in case of kubernetes we write apply plan in job result because we need to reference then in next plan cycle
	// this is not the case with other component types apply results are empty
	var apiRes *models.ServiceCreateRunnerJobExecutionResultRequest
	resultDisplay := map[string]interface{}{}
	resultContents, err := json.Marshal(kubernetesManifestPlan)
	json.Unmarshal(resultContents, &resultDisplay)
	if err != nil {
		h.writeErrorResult(ctx, err)
		l.Error("Failed to create API result", zap.Error(err), zap.String("jobID", job.ID))
		return fmt.Errorf("unable to create api result from release: %w", err)
	}
	apiRes, err = h.createAPIResult(rel, string(resultContents), resultDisplay)
	if err != nil {
		h.writeErrorResult(ctx, err)
		l.Error("Failed to create API result", zap.Error(err), zap.String("jobID", job.ID))
		return fmt.Errorf("unable to create api result from release: %w", err)
	}

	_, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, apiRes)
	if err != nil {
		l.Error("Failed to create job execution result", zap.Error(err), zap.String("jobID", job.ID))
		h.errRecorder.Record("write job execution result", err)
	}

	l.Debug("Exec function completed successfully", zap.String("jobID", job.ID))
	return nil
}

type operationOutput struct {
	Op   types.KubernetesManifestPlanOperation `json:"op,omitempty"`
	Name string                                `json:"name,omitempty"`
}

func (h *handler) execApply(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource) (*[]operationOutput, error) {
	output := make([]operationOutput, 0, len(resources))
	for _, resource := range resources {
		applyOptions := metav1.ApplyOptions{FieldManager: "kube-apply"}
		var err error
		if resource.namespaced {
			_, err = client.
				Resource(resource.groupVersionResource).
				Namespace(resource.namespace).
				Apply(ctx, resource.name, resource.obj, applyOptions)
		} else {
			_, err = client.
				Resource(resource.groupVersionResource).
				Apply(ctx, resource.name, resource.obj, applyOptions)
		}
		if err != nil {
			return &output, fmt.Errorf(
				"apply error for resource [Group: %s, Version: %s, Kind: %s, Namespace: %s, Name: %s]: %w",
				resource.groupVersionResource.Group,
				resource.groupVersionResource.Version,
				resource.groupVersionResource.Resource,
				resource.namespace,
				resource.name,
				err,
			)
		}
		output = append(output, operationOutput{
			Op:   types.KubernetesManifestPlanOperationApply,
			Name: resource.name,
		})
	}
	return &output, nil
}

func (h *handler) execDelete(ctx context.Context, client dynamic.Interface, resources []*kubernetesResource) (*[]operationOutput, error) {
	output := make([]operationOutput, 0, len(resources))
	for _, resource := range resources {
		deleteOptions := metav1.DeleteOptions{}
		var err error
		if resource.namespaced {
			err = client.Resource(resource.groupVersionResource).
				Namespace(resource.namespace).
				Delete(ctx, resource.name, deleteOptions)
		} else {
			err = client.Resource(resource.groupVersionResource).
				Delete(ctx, resource.name, deleteOptions)
		}
		if err != nil {
			return &output, fmt.Errorf(
				"delete error for resource [Group: %s, Version: %s, Kind: %s, Namespace: %s, Name: %s]: %w",
				resource.groupVersionResource.Group,
				resource.groupVersionResource.Version,
				resource.groupVersionResource.Resource,
				resource.namespace,
				resource.name,
				err,
			)
		}
		output = append(output, operationOutput{
			Op:   types.KubernetesManifestPlanOperationDelete,
			Name: resource.name,
		})
	}
	return &output, nil
}

func (h *handler) getPreviousConfigForResource(
	ctx context.Context,
	in *kubernetesResource,
	prev []*kubernetesResource,
) kubernetesResource {
	for _, r := range prev {
		if resourceName(r) == resourceName(in) {
			return *r
		}
	}
	return kubernetesResource{}
}
