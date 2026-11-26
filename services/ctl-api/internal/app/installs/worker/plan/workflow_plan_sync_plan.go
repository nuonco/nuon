package plan

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/registryprune"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// Placeholder identity for a sandboxed sync with no pullable source image.
const (
	fakeSyncRepository = "registry.example.com/nuon/app-service"
	fakeSyncTag        = "v1.2.3"
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

	return p.RenderSyncOCIPlan(l, &RenderSyncOCIPlanInput{
		Deploy:    deploy,
		CompBuild: compBuild,
		Install:   install,
		SrcCfg:    srcCfg,
		DstCfg:    dstCfg,
	})
}

// RenderSyncOCIPlanInput carries the already-loaded data a sync OCI plan is rendered from.
type RenderSyncOCIPlanInput struct {
	Deploy    *app.InstallDeploy
	CompBuild *app.ComponentBuild
	Install   *app.Install
	SrcCfg    *configs.OCIRegistryRepository
	DstCfg    *configs.OCIRegistryRepository
}

// RenderSyncOCIPlan renders a sync OCI plan from already-loaded inputs.
func (p *Planner) RenderSyncOCIPlan(l *zap.Logger, in *RenderSyncOCIPlanInput) (*plantypes.SyncOCIPlan, error) {
	// For image-type builds with source-identity recorded, copy the
	// artifact by digest and tag the install-registry copy with the
	// resolved tag instead of an internal ID. Fall back to installRegistryTag
	// for non-image components and image builds without source identity.
	srcTag := in.Deploy.ComponentBuildID
	if in.CompBuild.SourceDigest != "" {
		srcTag = in.CompBuild.SourceDigest
	}
	dstTag := registryprune.InstallRegistryTag(
		in.Deploy.ComponentName,
		installRegistryTag(in.Deploy),
		in.CompBuild.SourceDigest,
		in.CompBuild.ResolvedTag,
		in.DstCfg.RegistryType == configs.OCIRegistryTypeECR,
	)

	pln := &plantypes.SyncOCIPlan{
		Src:    in.SrcCfg,
		SrcTag: srcTag,

		DstTag: dstTag,
		Dst:    in.DstCfg,
	}

	if in.Install.SandboxMode.Bool {
		pln.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: sandboxSyncOutputs(in.CompBuild),
		}
	}

	return pln, nil
}

// sandboxSyncOutputs decides what a sandboxed image sync reports as the artifact
// it synced. A component sourcing a public image reports that image at its real
// tag, so the install's outputs name the artifact the component actually points
// at rather than a fabricated registry. Anything Nuon builds, and any source
// behind registry credentials the sandbox never mints, keeps the placeholder.
func sandboxSyncOutputs(build *app.ComponentBuild) map[string]any {
	cfg := build.ComponentConfigConnection.ExternalImageComponentConfig
	if cfg == nil || cfg.ImageURL == "" {
		return plantypes.FakeOCISyncOutputs(fakeSyncRepository, fakeSyncTag)
	}
	if cfg.AWSECRImageConfig != nil || cfg.GCPGARImageConfig != nil || cfg.AzureACRImageConfig != nil {
		return plantypes.FakeOCISyncOutputs(fakeSyncRepository, fakeSyncTag)
	}

	// A config carrying only an update_policy has no tag until a real build
	// resolves one against the source registry, which a sandbox never does.
	tag := generics.FirstNonEmptyString(build.ResolvedTag, cfg.Tag)
	if tag == "" {
		return plantypes.FakeOCISyncOutputs(fakeSyncRepository, fakeSyncTag)
	}

	return plantypes.FakePublicOCISyncOutputs(cfg.ImageURL, tag)
}
