package opentofu

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/kube/config"
	pkgopentofuremote "github.com/nuonco/nuon/pkg/opentofu/binary/remote"
	pkgopentofuworkspace "github.com/nuonco/nuon/pkg/opentofu/workspace"
	dirarchive "github.com/nuonco/nuon/pkg/terraform/archive/dir"
	httpbackend "github.com/nuonco/nuon/pkg/terraform/backend/http"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	"github.com/nuonco/nuon/pkg/terraform/hooks/noop"
	authvars "github.com/nuonco/nuon/pkg/terraform/variables/auth"
	staticvars "github.com/nuonco/nuon/pkg/terraform/variables/static"
	tfworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

// buildBinary always installs the OpenTofu CLI via tofudl. Unlike terraform,
// v1 has no build-vendored-binary convention for opentofu artifacts yet, so
// there is no bundled-binary detection to fall back around — the remote
// download is the only path.
func (p *handler) buildBinary(requestedVersion string) (binary.Binary, error) {
	return pkgopentofuremote.New(p.v, pkgopentofuremote.WithVersion(requestedVersion))
}

// detectAndLogMirror runs the shared filesystem-mirror detection over the
// unpacked artifact and emits a single Info log describing which
// provider-resolution path the install runner will take. The detection is
// purely filesystem-driven and tool-agnostic: the build runner vendors
// providers into the same conventional directory for both engines, so
// tfworkspace.DetectFilesystemMirror finds an opentofu mirror too. The
// returned value is suitable to pass to
// pkgopentofuworkspace.WithFilesystemMirror.
func (p *handler) detectAndLogMirror(archBase string) string {
	path := tfworkspace.DetectFilesystemMirror(archBase)
	platforms := tfworkspace.MirrorPlatforms(archBase)

	switch {
	case path != "":
		p.l.Info("opentofu: build-vendored provider mirror detected, using airgap resolution",
			zap.String("mirror_path", path),
			zap.Strings("mirror_platforms", platforms),
		)
	case len(platforms) > 0:
		p.l.Warn("opentofu: provider mirror present but does not include host platform; falling back to direct registry resolution",
			zap.Strings("mirror_platforms", platforms),
		)
	default:
		p.l.Info("opentofu: no provider mirror in artifact, using direct registry resolution")
	}
	return path
}

// GetWorkspace returns a valid workspace for working with this plugin
func (p *handler) GetWorkspace(ctx context.Context) (tfworkspace.Workspace, error) {
	arch, err := dirarchive.New(p.v,
		dirarchive.WithPath(p.state.arch.BasePath()),
		dirarchive.WithAddBackendFile("http"),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithIgnoreDotTerraformDir(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(p.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: p.cfg.RunnerAPIURL,
		WorkspaceID: p.state.plan.OpentofuDeployPlan.OpentofuBackend.WorkspaceID,
		Token:       p.cfg.APIToken(),
		JobID:       p.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := p.buildBinary(p.state.opentofuCfg.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	extraEnvVars := make(map[string]string, 0)
	if p.state.plan.OpentofuDeployPlan.ClusterInfo != nil {
		extraEnvVars[config.DefaultKubeConfigEnvVar] = config.DefaultKubeConfigFilename
		// The Terraform Kubernetes provider does not read the standard
		// KUBECONFIG env var — it uses KUBE_CONFIG_PATH instead.
		extraEnvVars["KUBE_CONFIG_PATH"] = config.DefaultKubeConfigFilename
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(p.state.plan.OpentofuDeployPlan.Vars),
		staticvars.WithFiles(p.state.plan.OpentofuDeployPlan.VarsFiles),
		staticvars.WithEnvVars(p.state.plan.OpentofuDeployPlan.EnvVars),
		staticvars.WithEnvVars(extraEnvVars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(p.v,
		authvars.WithAWSAuth(p.state.auth.AWSAuth),
		authvars.WithAzureAuth(p.state.auth.AzureAuth),
		authvars.WithGCPAuth(p.state.auth.GCPAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	hooks := noop.New()

	wkspace, err := pkgopentofuworkspace.New(p.v,
		pkgopentofuworkspace.WithHooks(hooks),
		pkgopentofuworkspace.WithArchive(arch),
		pkgopentofuworkspace.WithBackend(back),
		pkgopentofuworkspace.WithBinary(bin),
		pkgopentofuworkspace.WithVariables(vars),
		pkgopentofuworkspace.WithVariables(authVars),
		// Empty path = no-op; only enables the mirror if the build runner
		// actually shipped one inside the OCI artifact.
		pkgopentofuworkspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

// GetWorkspaceWithPlan returns a valid workspace for working with this plugin,
// seeded with the saved plan bytes from the plan phase.
func (p *handler) GetWorkspaceWithPlan(ctx context.Context, planBytes []byte) (tfworkspace.Workspace, error) {
	arch, err := dirarchive.New(p.v,
		dirarchive.WithPath(p.state.arch.BasePath()),
		dirarchive.WithAddBackendFile("http"),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithIgnoreDotTerraformDir(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(p.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: p.cfg.RunnerAPIURL,
		WorkspaceID: p.state.plan.OpentofuDeployPlan.OpentofuBackend.WorkspaceID,
		Token:       p.cfg.APIToken(),
		JobID:       p.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := p.buildBinary(p.state.opentofuCfg.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	extraEnvVars := make(map[string]string, 0)
	if p.state.plan.OpentofuDeployPlan.ClusterInfo != nil {
		extraEnvVars[config.DefaultKubeConfigEnvVar] = config.DefaultKubeConfigFilename
		// The Terraform Kubernetes provider does not read the standard
		// KUBECONFIG env var — it uses KUBE_CONFIG_PATH instead.
		extraEnvVars["KUBE_CONFIG_PATH"] = config.DefaultKubeConfigFilename
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(p.state.plan.OpentofuDeployPlan.Vars),
		staticvars.WithFiles(p.state.plan.OpentofuDeployPlan.VarsFiles),
		staticvars.WithEnvVars(p.state.plan.OpentofuDeployPlan.EnvVars),
		staticvars.WithEnvVars(extraEnvVars),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(p.v,
		authvars.WithAWSAuth(p.state.auth.AWSAuth),
		authvars.WithAzureAuth(p.state.auth.AzureAuth),
		authvars.WithGCPAuth(p.state.auth.GCPAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	hooks := noop.New()

	wkspace, err := pkgopentofuworkspace.New(p.v,
		pkgopentofuworkspace.WithHooks(hooks),
		pkgopentofuworkspace.WithArchive(arch),
		pkgopentofuworkspace.WithBackend(back),
		pkgopentofuworkspace.WithBinary(bin),
		pkgopentofuworkspace.WithVariables(vars),
		pkgopentofuworkspace.WithVariables(authVars),
		pkgopentofuworkspace.WithPlanBytes(planBytes),
		// See GetWorkspace for an explanation of the filesystem mirror.
		pkgopentofuworkspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
