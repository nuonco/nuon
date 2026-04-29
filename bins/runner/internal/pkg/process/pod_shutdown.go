package process

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/nuonco/nuon/bins/runner/internal"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/k8s"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/settings"
)

type podShutdown struct {
	cfg      *internal.Config
	settings *settings.Settings
	l        *zap.Logger
}

func newPodShutdown(cfg *internal.Config, settings *settings.Settings, l *zap.Logger) *podShutdown {
	if !cfg.DeletePodOnShutdown || cfg.PodName == "" || cfg.PodNamespace == "" {
		return nil
	}

	return &podShutdown{cfg: cfg, settings: settings, l: l}
}

func (ps *podShutdown) execute(ctx context.Context) error {
	clientset, _, _, err := k8s.ClientsetInCluster()
	if err != nil {
		return fmt.Errorf("unable to create in-cluster k8s client: %w", err)
	}

	if err := ps.updateDeploymentImage(ctx, clientset); err != nil {
		return fmt.Errorf("unable to update deployment image: %w", err)
	}

	if err := ps.deletePod(ctx, clientset); err != nil {
		return fmt.Errorf("unable to delete own pod: %w", err)
	}

	return nil
}

func (ps *podShutdown) deletePod(ctx context.Context, clientset *kubernetes.Clientset) error {
	ps.l.Info("deleting own pod",
		zap.String("pod_name", ps.cfg.PodName),
		zap.String("pod_namespace", ps.cfg.PodNamespace),
	)

	if err := clientset.CoreV1().Pods(ps.cfg.PodNamespace).Delete(ctx, ps.cfg.PodName, metav1.DeleteOptions{}); err != nil {
		return err
	}

	ps.l.Info("successfully deleted own pod")
	return nil
}

func (ps *podShutdown) updateDeploymentImage(ctx context.Context, clientset *kubernetes.Clientset) error {
	if ps.cfg.DeploymentName == "" {
		return nil
	}

	imageTag := ps.settings.ContainerImageTag
	imageURL := ps.settings.ContainerImageURL
	if imageTag == "" || imageURL == "" {
		return nil
	}

	image := fmt.Sprintf("%s:%s", imageURL, imageTag)
	ps.l.Info("updating deployment image",
		zap.String("deployment", ps.cfg.DeploymentName),
		zap.String("namespace", ps.cfg.PodNamespace),
		zap.String("image", image),
	)

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  ps.cfg.DeploymentName,
							"image": image,
						},
					},
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("unable to marshal deployment patch: %w", err)
	}

	if _, err := clientset.AppsV1().Deployments(ps.cfg.PodNamespace).Patch(
		ctx,
		ps.cfg.DeploymentName,
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	); err != nil {
		return err
	}

	ps.l.Info("successfully updated deployment image", zap.String("image", image))
	return nil
}
