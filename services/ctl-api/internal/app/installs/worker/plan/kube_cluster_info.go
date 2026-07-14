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

// resolveKubernetesContext returns the *kube.ClusterInfo a component should
// use during deploy planning.
//
// If the component declares a kubernetes_context, the cluster fields are
// templated from the named context's source peer component
// (`{{.nuon.components.<peer>.outputs.cluster.*}}`). The cloud-auth selection
// (AWS / Azure / GCP) still keys off install-level stack outputs, since a
// peer cluster lives inside the same install and is therefore on the same
// cloud as the sandbox.
//
// If the component does not declare a context, the resolver falls back to
// today's implicit sandbox-default behavior: when the sandbox emits cluster
// outputs the cluster fields are templated from the sandbox; when it
// doesn't (Lambda / ECS sandboxes, etc.) we return nil and the component
// runs without ClusterInfo.
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

// resolveKubernetesContextByName resolves a named kubernetes_context (or the
// sandbox default when name is empty) to a templated *kube.ClusterInfo. It is
// the shared core used by both component deploy planning and action workflow
// run planning; see resolveKubernetesContext for the full semantics.
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
		// no recognized cloud; nothing to wire up
		return nil, nil
	}

	if obj.GCPAuth != nil && obj.GCPAuth.ImpersonateServiceAccount != "" &&
		p.gcpRunnerInstanceRoleFallback(ctx, &stack.InstallStackOutputs) {
		l.Info("gcp-runner-instance-role enabled; using runner instance auth for cluster access")
		obj.GCPAuth = withoutGCPImpersonation(obj.GCPAuth)
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

// lookupKubernetesContextSource finds the named kubernetes_context on the
// AppConfig and returns the name of its source peer component. Errors when
// the named context is missing — config-time validation should have caught
// this in pkg/config/AppConfig.resolveKubernetesContexts, so a miss here
// indicates a stale config or out-of-band mutation.
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

// sandboxEmitsClusterOutputs reports whether the sandbox's outputs include a
// `cluster` key — the signal we use to decide whether the sandbox can act as
// the implicit default context.
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

// clusterInfoFromCloud builds a templated ClusterInfo for the install's cloud,
// where clusterPath is the dotted path (without surrounding `{{ }}`) at which
// the `cluster` output object lives in the rendered state — e.g.
// `.nuon.sandbox.outputs.cluster` or `.nuon.components.foo.outputs.cluster`.
//
// Returns nil if the install's stack outputs don't match a recognized cloud.
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

// assertCloudAuth mirrors today's checks that the cloudAuth side has the
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
