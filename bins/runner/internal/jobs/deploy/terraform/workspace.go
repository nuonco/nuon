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

// buildBinary picks between a build-vendored terraform CLI binary inside
// the OCI artifact and the existing remote (releases.hashicorp.com) path.
// The picker is purely filesystem-driven and feature-flag-unaware on the
// install side: workspace.DetectBundledBinary returns "" for any artifact
// that doesn't ship a usable host-platform binary at the requested
// version, in which case we fall through to remotebinary — the exact code
// path this runner has always taken. Old artifacts and old runners both
// hit the remote path with no behavior change.
func (p *handler) buildBinary(archBase, requestedVersion string) (binary.Binary, error) {
	if path := p.detectAndLogBundledBinary(archBase, requestedVersion); path != "" {
		return localbinary.New(p.v, localbinary.WithPath(path))
	}
	return remotebinary.New(p.v, remotebinary.WithVersion(requestedVersion))
}

// detectAndLogBundledBinary calls workspace.DetectBundledBinary and emits
// a single log line describing which terraform CLI path was chosen. We
// keep the detection + the log together so each install run produces
// exactly one operator-visible signal for the binary decision (mirroring
// detectAndLogMirror's contract for providers).
//
// Logged states:
//   - airgap-ready: bundled binary present, host-platform match, version
//     match → use it.
//   - version-mismatch: bundled binary present but VERSION sidecar
//     disagrees with requestedVersion → fall through to remote (warn).
//   - wrong-platform: bundled binaries present, none for this host → fall
//     through to remote (warn).
//   - absent: no bundled binary dir → silent, falls through. We don't
//     log the absent case to keep non-vendored installs noise-free; the
//     provider-mirror line already tells the operator whether the
//     artifact participates in airgap at all.
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

// detectAndLogMirror runs DetectFilesystemMirror and emits a single Info log
// describing which provider-resolution path the install runner will take.
// Returned value is suitable to pass to workspace.WithFilesystemMirror.
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

// GetWorkspace returns a valid workspace for working with this plugin
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
		// The Terraform Kubernetes provider does not read the standard
		// KUBECONFIG env var — it uses KUBE_CONFIG_PATH instead.
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
		// Empty path = no-op; only enables the mirror if the build runner
		// actually shipped one inside the OCI artifact.
		workspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

// GetWorkspace returns a valid workspace for working with this plugin
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
		// The Terraform Kubernetes provider does not read the standard
		// KUBECONFIG env var — it uses KUBE_CONFIG_PATH instead.
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
		// See GetWorkspace for an explanation of the filesystem mirror.
		workspace.WithFilesystemMirror(p.detectAndLogMirror(p.state.arch.BasePath())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
