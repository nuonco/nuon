package kubernetes_manifest

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-runner-go/models"
	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func (h *handler) Validate(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	l.Info("parsing kubernetes manifest to ensure correct")

	// Get the manifest content from the job
	manifestContent := h.state.plan.KubernetesManifestDeployPlan.Manifest
	if manifestContent == "" {
		return fmt.Errorf("no manifest content provided in job config")
	}

	// 1. YAML validation
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(manifestContent), &yamlData); err != nil {
		l.Error("failed to parse YAML manifest", zap.Error(err))
		return fmt.Errorf("invalid YAML format: %w", err)
	}
	l.Debug("YAML validation passed")

	// 2. Kubernetes manifest validation
	// Parse as unstructured Kubernetes object
	resources, err := h.getKubernetesResourcesFromManifest(h.state.kubeClient, manifestContent)
	if err != nil {
		l.Error("failed to parse Kubernetes manifest", zap.Error(err))
		return fmt.Errorf("invalid Kubernetes manifest format: %w", err)
	}

	for _, r := range resources {
		obj := r.obj

		// Validate required fields
		gvk := obj.GroupVersionKind()
		if gvk.Kind == "" {
			return fmt.Errorf("manifest missing required field: kind")
		}
		if gvk.Version == "" {
			return fmt.Errorf("manifest missing required field: apiVersion")
		}

		metadata := obj.Object["metadata"]
		if metadata == nil {
			return fmt.Errorf("manifest missing required field: metadata")
		}

		metadataMap, ok := metadata.(map[string]interface{})
		if !ok {
			return fmt.Errorf("manifest metadata field is not a valid object")
		}

		if name := obj.GetName(); name == "" {
			return fmt.Errorf("manifest missing required field: metadata.name")
		}

		// Set default namespace if not specified
		if namespace := obj.GetNamespace(); namespace == "" {
			metadataMap["namespace"] = h.state.plan.KubernetesManifestDeployPlan.Namespace
			l.Debug("set default namespace",
				zap.String("namespace", h.state.plan.KubernetesManifestDeployPlan.Namespace),
			)
		}
		l.Info("manifest validation passed",
			zap.String("kind", gvk.Kind),
			zap.String("name", obj.GetName()),
			zap.String("namespace", obj.GetNamespace()),
		)
	}
	l.Info("kubernetes manifest validation passed",
		zap.Int("resource_count", len(resources)),
	)

	return nil
}
