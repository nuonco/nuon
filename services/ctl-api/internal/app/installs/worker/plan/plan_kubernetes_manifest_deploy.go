package plan

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "embed"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	types "github.com/powertoolsdev/mono/pkg/types/components/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createKubernetesManifestDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.KubernetesManifestDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	state, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	// parse out various config fields
	cfg := compBuild.ComponentConfigConnection.KubernetesManifestComponentConfig
	if err := render.RenderStruct(cfg, stateData); err != nil {
		l.Error("error rendering helm config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	manifest := cfg.Manifest
	renderedManifest, err := render.RenderV2(manifest, stateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render")
	}

	// we might need namespace input in config or in manifest, if both are not present then it will go in default
	// namespace

	namespace := cfg.Namespace
	renderedNamespace, err := render.RenderV2(namespace, stateData)
	if err != nil {
		l.Error("error rendering namespace",
			zap.String("namespace", namespace),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
	}

	return &plantypes.KubernetesManifestDeployPlan{
		ClusterInfo: clusterInfo,
		Manifest:    renderedManifest,
		Namespace:   renderedNamespace,
	}, nil
}

func (p *Planner) createKubernetesManifestDeployPlanSandboxMode(req *plantypes.KubernetesManifestDeployPlan) (*plantypes.KubernetesSandboxMode, error) {
	obj := types.KubernetesManifestPlanContents{
		Plan: []*types.KubernetesManifestDiff{
			{
				GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "configmaps",
				},
				Namespace: "sandbox-namespace",
				Name:      "sandbox-configmap",
				Op:        types.KubernetesManifestPlanOperationApply,
				Before:    "",
				After: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sandbox-configmap
  namespace: sandbox-namespace
data:
  key1: value1
  key2: value2
`,
			},
			{
				GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "secrets",
				},
				Namespace: "sandbox-namespace",
				Name:      "sandbox-secret",
				Op:        types.KubernetesManifestPlanOperationDelete,
				Before: `
apiVersion: v1
kind: Secret
metadata:
  name: sandbox-secret
  namespace: sandbox-namespace
data:
  password: c2VjcmV0
`,
				After: "",
			},
			{
				GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "",
					Version:  "v1",
					Resource: "configmaps",
				},
				Namespace: "sandbox-namespace",
				Name:      "sandbox-configmap",
				Op:        types.KubernetesManifestPlanOperationApply,
				Before: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sandbox-configmap
  namespace: sandbox-namespace
data:
  key1: value1
  key2: value2
`,
				After: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sandbox-configmap
  namespace: sandbox-namespace
data:
  key1: updated-value1
  key2: value2
  key3: new-value3
`,
			},
		},
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal kubernetes manifest plan contents: %w", err)
	}
	return &plantypes.KubernetesSandboxMode{
		PlanContents:        string(b),
		PlanDisplayContents: string(b),
	}, nil
}
