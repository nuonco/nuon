package terraform

import (
	"context"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/kube/healthcheck"
)

const noopHealthCheckTimeout = 30 * time.Second

var workloadKinds = map[string]bool{
	"Deployment":  true,
	"StatefulSet": true,
	"DaemonSet":   true,
}

var terraformTypeToKind = map[string]string{
	"kubernetes_deployment":   "Deployment",
	"kubernetes_stateful_set": "StatefulSet",
	"kubernetes_daemon_set":   "DaemonSet",
}

func (p *handler) checkWorkloadHealthOnNoop(ctx context.Context, l *zap.Logger, plan *tfjson.Plan, state *tfjson.State) error {
	if !isTerraformPlanNoop(plan) {
		return nil
	}

	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		l.Debug("no terraform state available, skipping workload health check")
		return nil
	}

	resources := extractWorkloadResources(state.Values.RootModule)
	if len(resources) == 0 {
		l.Debug("no workload resources found in terraform state, skipping health check")
		return nil
	}

	l.Info("terraform plan has no changes, checking workload health",
		zap.Int("workload_count", len(resources)))

	kubeCfg, err := p.getKubeConfig(ctx)
	if err != nil {
		l.Warn("unable to get kube config for health check, skipping", zap.Error(err))
		return nil
	}

	return healthcheck.CheckWorkloadHealth(ctx, l, kubeCfg, resources, noopHealthCheckTimeout)
}

func (p *handler) getKubeConfig(ctx context.Context) (*rest.Config, error) {
	if p.state.plan.TerraformDeployPlan.ClusterInfo != nil {
		return kube.ConfigForCluster(ctx, p.state.plan.TerraformDeployPlan.ClusterInfo)
	}
	return kube.GetKubeConfig()
}

func isTerraformPlanNoop(plan *tfjson.Plan) bool {
	if plan == nil {
		return false
	}

	for _, rc := range plan.ResourceChanges {
		if rc.Change != nil {
			for _, action := range rc.Change.Actions {
				if action != tfjson.ActionNoop {
					return false
				}
			}
		}
	}

	for _, oc := range plan.OutputChanges {
		if oc != nil {
			for _, action := range oc.Actions {
				if action != tfjson.ActionNoop {
					return false
				}
			}
		}
	}

	return true
}

func extractWorkloadResources(module *tfjson.StateModule) []healthcheck.WorkloadResource {
	var resources []healthcheck.WorkloadResource

	for _, res := range module.Resources {
		if wr, ok := parseWorkloadResource(res); ok {
			resources = append(resources, wr)
		}
	}

	for _, child := range module.ChildModules {
		resources = append(resources, extractWorkloadResources(child)...)
	}

	return resources
}

func parseWorkloadResource(res *tfjson.StateResource) (healthcheck.WorkloadResource, bool) {
	if res.Type == "kubectl_manifest" {
		return parseKubectlManifestResource(res)
	}

	if kind, ok := terraformTypeToKind[res.Type]; ok {
		return parseHashicorpK8sResource(res, kind)
	}

	return healthcheck.WorkloadResource{}, false
}

func parseKubectlManifestResource(res *tfjson.StateResource) (healthcheck.WorkloadResource, bool) {
	attrs := res.AttributeValues
	if attrs == nil {
		return healthcheck.WorkloadResource{}, false
	}

	kind, _ := attrs["kind"].(string)
	if !workloadKinds[kind] {
		return healthcheck.WorkloadResource{}, false
	}

	name, _ := attrs["name"].(string)
	namespace, _ := attrs["namespace"].(string)

	if name == "" {
		return healthcheck.WorkloadResource{}, false
	}
	if namespace == "" {
		namespace = "default"
	}

	return healthcheck.WorkloadResource{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}, true
}

func parseHashicorpK8sResource(res *tfjson.StateResource, kind string) (healthcheck.WorkloadResource, bool) {
	attrs := res.AttributeValues
	if attrs == nil {
		return healthcheck.WorkloadResource{}, false
	}

	name, namespace := extractHashicorpK8sMetadata(attrs)
	if name == "" {
		return healthcheck.WorkloadResource{}, false
	}
	if namespace == "" {
		namespace = "default"
	}

	return healthcheck.WorkloadResource{
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}, true
}

// The hashicorp/kubernetes provider stores metadata as:
// {"metadata": [{"name": "x", "namespace": "y"}]}
func extractHashicorpK8sMetadata(attrs map[string]interface{}) (name, namespace string) {
	metadataRaw, ok := attrs["metadata"]
	if !ok {
		return "", ""
	}

	metadataList, ok := metadataRaw.([]interface{})
	if !ok || len(metadataList) == 0 {
		return "", ""
	}

	metadata, ok := metadataList[0].(map[string]interface{})
	if !ok {
		return "", ""
	}

	name, _ = metadata["name"].(string)
	namespace, _ = metadata["namespace"].(string)
	return name, namespace
}
