package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) createSyncPlan(ctx workflow.Context, req *CreateSyncPlanRequest) (*plantypes.SyncOCIPlan, error) {
	deploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	srcCfg, err := p.getOrgRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org registry repository")
	}

	dstCfg, err := p.getInstallRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository")
	}

	pln := &plantypes.SyncOCIPlan{
		Src:    srcCfg,
		SrcTag: deploy.ComponentBuildID,

		DstTag: deploy.ID,
		Dst:    dstCfg,
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, deploy.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
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
