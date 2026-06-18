package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
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

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	roleSelection, _, err := p.getRoleForDeploy(ctx, l, appCfg, deploy, compBuild, stack, installState)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get role for deploy")
	}

	dstCfg, err := p.getInstallRegistryRepositoryConfig(ctx, deploy, compBuild, appCfg, stack, installState, roleSelection)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository")
	}

	// For image-type builds with source-identity recorded, copy the
	// artifact by digest and tag the install-registry copy with the
	// resolved tag instead of an internal ID. Fall back to build/deploy ID
	// tagging for non-image components and image builds without source
	// identity.
	srcTag := deploy.ComponentBuildID
	dstTag := deploy.ID
	if compBuild.SourceDigest != "" {
		// oras.Copy resolves both tags and digest references, so passing
		// the manifest digest as the source ref pulls the exact same
		// content the runner originally resolved during build.
		srcTag = compBuild.SourceDigest

		// Push the install-registry copy under a tag the customer can
		// recognise. Prefer the tag the runner resolved (e.g. "1.25.5"),
		// and fall back to the build ID when the user pinned by digest
		// and there is no resolved tag. The digest is the canonical
		// identity of the artifact — the tag is metadata only and
		// duplicate tags across builds (e.g. successive deploys of
		// "1.25.5") are intentionally idempotent.
		dstTag = compBuild.ResolvedTag
		if dstTag == "" {
			dstTag = compBuild.ID
		} else if dstCfg.RegistryType == configs.OCIRegistryTypeECR {
			// ECR keeps one shared repo (repos must be pre-created), so
			// prefix the tag to keep resolved versions from colliding
			// across components.
			dstTag = imageNameSegment(deploy.ComponentName) + "-" + dstTag
		}
	}

	pln := &plantypes.SyncOCIPlan{
		Src:    srcCfg,
		SrcTag: srcTag,

		DstTag: dstTag,
		Dst:    dstCfg,
	}

	if install.SandboxMode.Bool {
		pln.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: map[string]any{
				"image": map[string]interface{}{
					// Sandbox outputs mirror the live runner sync outputs
					// shape: `repository` and `tag` are the bare repo and
					// resolved tag that user templates compose as
					// `{{.repository}}:{{.tag}}`; `ref` is the additive
					// digest-pinned form; `display_tag` carries the
					// human-friendly tag.
					"repository":    "registry.example.com/nuon/app-service",
					"tag":           "v1.2.3",
					"ref":           "registry.example.com/nuon/app-service@sha256:a123b456c789d012e345f678g901h234i567j890k123l456m789n012o345p",
					"display_tag":   "v1.2.3",
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
