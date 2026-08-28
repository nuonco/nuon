package plan

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/diff"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	types "github.com/nuonco/nuon/pkg/types/approvals"
	statepkg "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createKubernetesManifestDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *statepkg.State,
	installDeploy *app.InstallDeploy,
	roleSelection *operationroles.RoleSelection,
) (*plantypes.KubernetesManifestDeployPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(
		ctx,
		installDeploy.ComponentBuildID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	return p.RenderKubernetesManifestDeployPlan(l, &RenderKubernetesManifestDeployPlanInput{
		Stack:         stack,
		State:         state,
		StateData:     stateData,
		InstallDeploy: installDeploy,
		CompBuild:     compBuild,
		RoleSelection: roleSelection,
		ResolveClusterInfo: func(cloudAuth *CloudAuth) (*kube.ClusterInfo, error) {
			return p.resolveKubernetesContext(ctx, &compBuild.ComponentConfigConnection, appCfg, stack, state, cloudAuth)
		},
	})
}

// RenderKubernetesManifestDeployPlanInput carries the already-loaded data a
// kubernetes manifest deploy plan is rendered from.
type RenderKubernetesManifestDeployPlanInput struct {
	Stack         *app.InstallStack
	State         *statepkg.State
	StateData     map[string]any
	InstallDeploy *app.InstallDeploy
	CompBuild     *app.ComponentBuild
	RoleSelection *operationroles.RoleSelection

	// ResolveClusterInfo resolves the kubernetes cluster at this exact point in
	// the plan render.
	ResolveClusterInfo func(*CloudAuth) (*kube.ClusterInfo, error)
}

// RenderKubernetesManifestDeployPlan is the pure core of
// createKubernetesManifestDeployPlan.
func (p *Planner) RenderKubernetesManifestDeployPlan(
	l *zap.Logger,
	in *RenderKubernetesManifestDeployPlanInput,
) (*plantypes.KubernetesManifestDeployPlan, error) {
	// parse out various config fields
	cfg := in.CompBuild.ComponentConfigConnection.KubernetesManifestComponentConfig
	if err := render.RenderStruct(cfg, in.StateData); err != nil {
		l.Error("error rendering kubernetes manifest config",
			zap.Error(err),
			zap.Any("state", in.StateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	// Render namespace with install state - namespace supports template variables like {{.nuon.install.id}}
	namespace := cfg.Namespace
	renderedNamespace, err := render.RenderV2(namespace, in.StateData)
	if err != nil {
		l.Error("error rendering namespace",
			zap.String("namespace", namespace),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	manifest := cfg.Manifest
	renderedManifest, err := render.RenderV2(manifest, in.StateData)
	if err != nil {
		l.Error("error rendering manifest",
			zap.String("manifest", manifest),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	// Build OCI artifact reference from the install deploy's synced artifact
	// The manifest content is pulled from this artifact at runtime by the runner
	ociArtifact := in.InstallDeploy.OCIArtifact
	if ociArtifact.Repository == "" {
		return nil, errors.New("OCI artifact not found on install deploy - sync job may not have completed")
	}

	l.Info("using OCI artifact for kubernetes manifest deploy",
		zap.String("repository", ociArtifact.Repository),
		zap.String("tag", ociArtifact.Tag),
		zap.String("digest", ociArtifact.Digest))

	cloudAuth, err := p.AuthForDeploy(
		l,
		in.RoleSelection,
		in.Stack,
		fmt.Sprintf("component-deploy-%s", in.InstallDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := in.ResolveClusterInfo(cloudAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve kubernetes context")
	}

	// Ship the install state to the runner only when the manifest will be
	// loaded from the OCI artifact (kustomize path — inline manifests are
	// already pre-rendered above and the runner short-circuits the OCI pull
	// when plan.Manifest is non-empty). This lets the runner interpolate
	// {{.nuon.*}} placeholders that survived kustomize unchanged.
	var planState *statepkg.State
	if renderedManifest == "" {
		planState = in.State
	}

	return &plantypes.KubernetesManifestDeployPlan{
		ClusterInfo: clusterInfo,
		Namespace:   renderedNamespace,
		Manifest:    renderedManifest,
		OCIArtifact: &plantypes.OCIArtifactReference{
			URL:    ociArtifact.Repository,
			Tag:    ociArtifact.Tag,
			Digest: ociArtifact.Digest,
		},
		State:     planState,
		AWSAuth:   cloudAuth.AWS,
		AzureAuth: cloudAuth.Azure,
		GCPAuth:   cloudAuth.GCP,
	}, nil
}

func (p *Planner) createKubernetesManifestDeployPlanSandboxMode(
	req *plantypes.KubernetesManifestDeployPlan,
) (*plantypes.KubernetesSandboxMode, error) {
	obj := types.KubernetesManifestPlanContents{
		Plan: "{\n  \"diff\": [\n    {\n      \"_version\": \"2\",\n      \"name\": \"demo\",\n      \"namespace\": \"default\",\n      \"kind\": \"ConfigMap\",\n      \"api\": \"/v1\",\n      \"resource\": \"configmaps\",\n      \"op\": \"apply\",\n      \"type\": 3,\n      \"dry_run\": true,\n      \"entries\": [\n        {\n          \"path\": \"data.sample_data\",\n          \"original\": \"3\",\n          \"applied\": \"4\",\n          \"type\": 3,\n          \"payload\": \"  map[string]any{\\n  \\t\\\"apiVersion\\\": string(\\\"v1\\\"),\\n- \\t\\\"data\\\":       map[string]any{\\\"sample_data\\\": string(\\\"3\\\")},\\n+ \\t\\\"data\\\":       map[string]any{\\\"sample_data\\\": string(\\\"4\\\")},\\n  \\t\\\"kind\\\":       string(\\\"ConfigMap\\\"),\\n  \\t\\\"metadata\\\":   map[string]any{\\\"name\\\": string(\\\"demo\\\"), ...},\\n  }\\n\"\n        }\n      ]\n    }\n  ]\n}",
		Op:   "apply",
		ContentDiff: []diff.ResourceDiff{
			{
				Version:   "2",
				Name:      "demo",
				Namespace: "default",
				Kind:      "ConfigMap",
				ApiPath:   "/v1",
				Resource:  "configmaps",
				Operation: "apply",
				Type:      3,
				DryRun:    true,
				Entries: []diff.DiffEntry{
					{
						Path:     "data.sample_data",
						Original: "3",
						Applied:  "4",
						Type:     3,
						Payload:  "  map[string]any{\n  \t\"apiVersion\": string(\"v1\"),\n- \t\"data\":       map[string]any{\"sample_data\": string(\"3\")},\n+ \t\"data\":       map[string]any{\"sample_data\": string(\"4\")},\n  \t\"kind\":       string(\"ConfigMap\"),\n  \t\"metadata\":   map[string]any{\"name\": string(\"demo\"), ...},\n  }\n",
					},
				},
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
