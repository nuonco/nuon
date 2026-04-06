package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) createSyncPlan(ctx workflow.Context, req *CreateSyncPlanRequest) (*plantypes.SyncOCIPlan, error) {
	deploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	compBuild, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, deploy.ComponentBuildID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component build")
	}

	srcCfg, err := p.getOrgRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org registry repository")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	dstCfg, err := p.getInstallRegistryRepositoryConfig(ctx, deploy, compBuild, appCfg, stack, installState)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository")
	}

	pln := &plantypes.SyncOCIPlan{
		Src:    srcCfg,
		SrcTag: deploy.ComponentBuildID,

		DstTag: deploy.ID,
		Dst:    dstCfg,
	}

	if install.SandboxMode.Bool {
		pln.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: map[string]any{
				"image": map[string]interface{}{
					"tag":           "v1.2.3",
					"repository":    "nuon/app-service",
					"media_type":    "application/vnd.docker.distribution.manifest.v2+json",
					"digest":        "sha256:a123b456c789d012e345f678g901h234i567j890k123l456m789n012o345p",
					"size":          28437192,
					"urls":          []string{"registry.example.com/nuon/app-service:v1.2.3"},
					"annotations":   map[string]string{"org.opencontainers.image.created": "2024-04-29T10:15:30Z"},
					"artifact_type": "application/vnd.docker.container.image.v1+json",
					"platform": map[string]any{
						"architecture": "arm64",
						"os":           "linux",
						"os_version":   "10.0",
						"variant":      "v8",
						"os_features":  []string{"sse4", "aes"},
					},
				},
			},
		}
	}

	return pln, nil
}
