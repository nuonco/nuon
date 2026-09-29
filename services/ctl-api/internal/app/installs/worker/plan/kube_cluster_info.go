package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) resolveKubernetesContext(
	ctx workflow.Context,
	componentConfig *app.ComponentConfigConnection,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	cloudAuth *CloudAuth,
) (*kube.ClusterInfo, error) {
	contextName := ""
	if componentConfig != nil {
		contextName = componentConfig.KubernetesContextName
	}
	return p.resolveKubernetesContextByName(ctx, contextName, appCfg, stack, state, cloudAuth)
}

func (p *Planner) resolveKubernetesContextByName(
	ctx workflow.Context,
	contextName string,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	cloudAuth *CloudAuth,
) (*kube.ClusterInfo, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state data")
	}

	var (
		clusterPath string
		obj         *kube.ClusterInfo
	)

	if contextName != "" {
		peer, err := lookupKubernetesContextSource(appCfg, contextName)
		if err != nil {
			return nil, err
		}
		l.Info("resolving kubernetes_context",
			zap.String("context", contextName),
			zap.String("source_component", peer),
		)
		clusterPath = fmt.Sprintf(".nuon.components.%s.outputs.cluster", peer)
		obj = clusterInfoFromCloud(stack, cloudAuth, clusterPath)
	} else {
		l.Info("no kubernetes_context declared; falling back to sandbox default")
		if !sandboxEmitsClusterOutputs(stateData) {
			l.Info("sandbox outputs do not include kubernetes cluster info, skipping")
			return nil, nil
		}
		clusterPath = ".nuon.sandbox.outputs.cluster"
		obj = clusterInfoFromCloud(stack, cloudAuth, clusterPath)
	}

	if obj == nil {
		return nil, nil
	}

	if err := assertCloudAuth(stack, cloudAuth); err != nil {
		return nil, err
	}

	if err := render.RenderStruct(obj, stateData); err != nil {
		l.Error("error rendering cluster info",
			zap.Any("cluster-info", obj),
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	l.Info("successfully resolved kubernetes context, including in plan")
	return obj, nil
}

func lookupKubernetesContextSource(appCfg *app.AppConfig, name string) (string, error) {
	if appCfg == nil {
		return "", errors.Errorf("kubernetes_context %q referenced but app config is nil", name)
	}
	for _, c := range appCfg.KubernetesContextsConfig.Contexts {
		if c.Name == name {
			if c.SourceComponentName == "" {
				return "", errors.Errorf("kubernetes_context %q has no source component", name)
			}
			return c.SourceComponentName, nil
		}
	}
	return "", errors.Errorf("kubernetes_context %q is not defined on this app config", name)
}

func sandboxEmitsClusterOutputs(stateData map[string]any) bool {
	sandbox, ok := stateData["sandbox"].(map[string]any)
	if !ok {
		return false
	}
	outputs, ok := sandbox["outputs"].(map[string]any)
	if !ok {
		return false
	}
	cluster, ok := outputs["cluster"]
	return ok && cluster != nil
}

func clusterInfoFromCloud(stack *app.InstallStack, cloudAuth *CloudAuth, clusterPath string) *kube.ClusterInfo {
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:       fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint: fmt.Sprintf("{{%s.endpoint}}", clusterPath),
			CAData:   fmt.Sprintf("{{%s.certificate_authority_data}}", clusterPath),
			AWSAuth:  cloudAuth.AWS,
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:        fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint:  fmt.Sprintf("{{%s.host}}", clusterPath),
			CAData:    fmt.Sprintf("{{%s.cluster_ca_certificate}}", clusterPath),
			AzureAuth: cloudAuth.Azure,
		}
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:       fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint: fmt.Sprintf("{{%s.endpoint}}", clusterPath),
			CAData:   fmt.Sprintf("{{%s.certificate_authority_data}}", clusterPath),
			GCPAuth:  cloudAuth.GCP,
		}
	}
	return nil
}

// why: assertCloudAuth mirrors today's checks that the cloudAuth side has the
// credentials matching the install's cloud. (GCP intentionally skipped to
// preserve prior behavior — the original code only required AWS/Azure auth.)
func assertCloudAuth(stack *app.InstallStack, cloudAuth *CloudAuth) error {
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		if cloudAuth.AWS == nil {
			return errors.New("aws auth information not provided")
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		if cloudAuth.Azure == nil {
			return errors.New("azure auth information not provided")
		}
	}
	return nil
}
