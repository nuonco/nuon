package plan

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	_ "embed"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/diff"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	types "github.com/nuonco/nuon/pkg/types/components/plan"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createKubernetesManifestDeployPlan(
	ctx workflow.Context,
	req *CreateDeployPlanRequest,
	appCfg *app.AppConfig,
	stack *app.InstallStack,
	state *state.State,
	installDeploy *app.InstallDeploy,
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

	// parse out various config fields
	cfg := compBuild.ComponentConfigConnection.KubernetesManifestComponentConfig
	if err := render.RenderStruct(cfg, stateData); err != nil {
		l.Error("error rendering kubernetes manifest config",
			zap.Error(err),
			zap.Any("state", stateData),
		)
		return nil, errors.Wrap(err, "unable to render config")
	}

	// Render namespace with install state - namespace supports template variables like {{.nuon.install.id}}
	namespace := cfg.Namespace
	renderedNamespace, err := render.RenderV2(namespace, stateData)
	if err != nil {
		l.Error("error rendering namespace",
			zap.String("namespace", namespace),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	manifest := cfg.Manifest
	renderedManifest, err := render.RenderV2(manifest, stateData)
	if err != nil {
		l.Error("error rendering manifest",
			zap.String("manifest", manifest),
			zap.Error(err))
		return nil, errors.Wrap(err, "unable to render namespace")
	}

	// Build OCI artifact reference from the install deploy's synced artifact
	// The manifest content is pulled from this artifact at runtime by the runner
	ociArtifact := installDeploy.OCIArtifact
	if ociArtifact.Repository == "" {
		return nil, errors.New("OCI artifact not found on install deploy - sync job may not have completed")
	}

	l.Info("using OCI artifact for kubernetes manifest deploy",
		zap.String("repository", ociArtifact.Repository),
		zap.String("tag", ociArtifact.Tag),
		zap.String("digest", ociArtifact.Digest))

	cloudAuth, err := p.getAuthForDeploy(
		ctx,
		installDeploy,
		compBuild,
		appCfg,
		stack,
		state,
		fmt.Sprintf("component-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for deploy")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster info")
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
