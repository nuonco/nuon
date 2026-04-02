package kubernetes_manifest

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/kube/healthcheck"
)

const k8sManifestNoopHealthCheckTimeout = 30 * time.Second

var workloadKinds = map[string]bool{
	"Deployment":  true,
	"StatefulSet": true,
	"DaemonSet":   true,
}

func (h *handler) checkWorkloadHealthOnNoop(ctx context.Context, l *zap.Logger, desiredResources []*kubernetesResource) error {
	var workloads []healthcheck.WorkloadResource

	for _, res := range desiredResources {
		if !workloadKinds[res.groupVersionKind.Kind] {
			continue
		}

		ns := res.namespace
		if ns == "" {
			ns = "default"
		}

		workloads = append(workloads, healthcheck.WorkloadResource{
			Kind:      res.groupVersionKind.Kind,
			Name:      res.name,
			Namespace: ns,
		})
	}

	if len(workloads) == 0 {
		return nil
	}

	l.Info("plan has no changes, checking workload health",
		zap.Int("workload_count", len(workloads)))

	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.plan.KubernetesManifestDeployPlan.ClusterInfo)
	if err != nil {
		l.Warn("unable to get kube config for health check, skipping", zap.Error(err))
		return nil
	}

	return healthcheck.CheckWorkloadHealth(ctx, l, kubeCfg, workloads, k8sManifestNoopHealthCheckTimeout)
}
