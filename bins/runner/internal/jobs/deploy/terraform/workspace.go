package terraform

import (
	"context"
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/kube/config"
	dirarchive "github.com/nuonco/nuon/pkg/terraform/archive/dir"
	httpbackend "github.com/nuonco/nuon/pkg/terraform/backend/http"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	localbinary "github.com/nuonco/nuon/pkg/terraform/binary/local"
	remotebinary "github.com/nuonco/nuon/pkg/terraform/binary/remote"
	"github.com/nuonco/nuon/pkg/terraform/hooks/noop"
	authvars "github.com/nuonco/nuon/pkg/terraform/variables/auth"
	staticvars "github.com/nuonco/nuon/pkg/terraform/variables/static"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

func (p *handler) buildBinary(archBase, requestedVersion string) (binary.Binary, error) {
	if path := p.detectAndLogBundledBinary(archBase, requestedVersion); path != "" {
		return localbinary.New(p.v, localbinary.WithPath(path))
	}
	return remotebinary.New(p.v, remotebinary.WithVersion(requestedVersion))
}

func (p *handler) detectAndLogBundledBinary(archBase, requestedVersion string) string {
	path := workspace.DetectBundledBinary(archBase, requestedVersion)
	bundledVersion := workspace.BundledBinaryVersion(archBase)
	bundledPlatforms := workspace.BundledBinaryPlatforms(archBase)
	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	switch {
	case path != "":
		p.l.Info("terraform: build-vendored CLI binary detected, using airgap binary",
			zap.String("arch_base", archBase),
			zap.String("bundled_binary_path", path),
			zap.String("host_platform", hostPlatform),
			zap.String("bundled_version", bundledVersion),
			zap.String("requested_version", requestedVersion),
			zap.Strings("bundled_platforms", bundledPlatforms),
		)
	case len(bundledPlatforms) > 0 && bundledVersion != "" && bundledVersion != requestedVersion:
		p.l.Warn("terraform: bundled CLI binary version mismatch; falling back to remote install",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.String("bundled_version", bundledVersion),
			zap.String("requested_version", requestedVersion),
			zap.Strings("bundled_platforms", bundledPlatforms),
		)
	case len(bundledPlatforms) > 0:
		p.l.Warn("terraform: bundled CLI binary present but does not include host platform; falling back to remote install",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.Strings("bundled_platforms", bundledPlatforms),
		)
	}
	return path
}

func (p *handler) detectAndLogMirror(archBase string) string {
	path := workspace.DetectFilesystemMirror(archBase)
	platforms := workspace.MirrorPlatforms(archBase)
	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	switch {
	case path != "":
		p.l.Info("terraform: build-vendored provider mirror detected, using airgap resolution",
			zap.String("arch_base", archBase),
			zap.String("mirror_path", path),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	case len(platforms) > 0:
		p.l.Warn("terraform: provider mirror present but does not include host platform; falling back to direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	default:
		p.l.Info("terraform: no provider mirror in artifact, using direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
		)
	}
	return path
}

func (p *handler) GetWorkspace(ctx context.Context) (workspace.Workspace, error) {
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
		WorkspaceID: p.state.plan.TerraformDeployPlan.TerraformBackend.WorkspaceID,
		Token:       p.cfg.RunnerAPIToken,
		JobID:       p.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := p.buildBinary(p.state.arch.BasePath(), p.state.terraformCfg.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	extraEnvVars := make(map[string]string, 0)
	if p.state.plan.TerraformDeployPlan.ClusterInfo != nil {
		extraEnvVars[config.DefaultKubeConfigEnvVar] = config.DefaultKubeConfigFilename
		extraEnvVars["KUBE_CONFIG_PATH"] = config.DefaultKubeConfigFilename
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(p.state.plan.TerraformDeployPlan.Vars),
		staticvars.WithFiles(p.state.plan.TerraformDeployPlan.VarsFiles),
		staticvars.WithEnvVars(p.state.plan.TerraformDeployPlan.EnvVars),
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

	wkspace, err := workspace.New(p.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		workspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

func (p *handler) GetWorkspaceWithPlan(ctx context.Context, planBytes []byte) (workspace.Workspace, error) {
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
		WorkspaceID: p.state.plan.TerraformDeployPlan.TerraformBackend.WorkspaceID,
		Token:       p.cfg.RunnerAPIToken,
		JobID:       p.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := p.buildBinary(p.state.arch.BasePath(), p.state.terraformCfg.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	extraEnvVars := make(map[string]string, 0)
	if p.state.plan.TerraformDeployPlan.ClusterInfo != nil {
		extraEnvVars[config.DefaultKubeConfigEnvVar] = config.DefaultKubeConfigFilename
		extraEnvVars["KUBE_CONFIG_PATH"] = config.DefaultKubeConfigFilename
	}

	vars, err := staticvars.New(p.v,
		staticvars.WithFileVars(p.state.plan.TerraformDeployPlan.Vars),
		staticvars.WithFiles(p.state.plan.TerraformDeployPlan.VarsFiles),
		staticvars.WithEnvVars(p.state.plan.TerraformDeployPlan.EnvVars),
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

	wkspace, err := workspace.New(p.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		workspace.WithPlanBytes(planBytes),
		workspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
