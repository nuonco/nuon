package terraform

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	dirarchive "github.com/nuonco/nuon/pkg/terraform/archive/dir"
	httpbackend "github.com/nuonco/nuon/pkg/terraform/backend/http"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	localbinary "github.com/nuonco/nuon/pkg/terraform/binary/local"
	remotebinary "github.com/nuonco/nuon/pkg/terraform/binary/remote"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
	"github.com/nuonco/nuon/pkg/terraform/hooks/noop"
	"github.com/nuonco/nuon/pkg/terraform/hooks/shell"
	authvars "github.com/nuonco/nuon/pkg/terraform/variables/auth"
	staticvars "github.com/nuonco/nuon/pkg/terraform/variables/static"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

// buildBinary picks between a build-vendored terraform CLI binary inside
// the OCI artifact and the existing remote (releases.hashicorp.com) path.
// See the deploy handler's buildBinary for the full contract; the sandbox
// handler shares the same compat semantics.
func (h *handler) buildBinary(archBase, requestedVersion string) (binary.Binary, error) {
	if path := h.detectAndLogBundledBinary(archBase, requestedVersion); path != "" {
		return localbinary.New(h.v, localbinary.WithPath(path))
	}
	return remotebinary.New(h.v, remotebinary.WithVersion(requestedVersion))
}

// detectAndLogBundledBinary mirrors the deploy handler's helper of the
// same name. Kept as a per-handler method so each handler can use its
// own *zap.Logger (h.l vs p.l) without plumbing a logger through the
// workspace package.
func (h *handler) detectAndLogBundledBinary(archBase, requestedVersion string) string {
	path := workspace.DetectBundledBinary(archBase, requestedVersion)
	bundledVersion := workspace.BundledBinaryVersion(archBase)
	bundledPlatforms := workspace.BundledBinaryPlatforms(archBase)
	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	switch {
	case path != "":
		h.l.Info("terraform: build-vendored CLI binary detected, using airgap binary",
			zap.String("arch_base", archBase),
			zap.String("bundled_binary_path", path),
			zap.String("host_platform", hostPlatform),
			zap.String("bundled_version", bundledVersion),
			zap.String("requested_version", requestedVersion),
			zap.Strings("bundled_platforms", bundledPlatforms),
		)
	case len(bundledPlatforms) > 0 && bundledVersion != "" && bundledVersion != requestedVersion:
		h.l.Warn("terraform: bundled CLI binary version mismatch; falling back to remote install",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.String("bundled_version", bundledVersion),
			zap.String("requested_version", requestedVersion),
			zap.Strings("bundled_platforms", bundledPlatforms),
		)
	case len(bundledPlatforms) > 0:
		h.l.Warn("terraform: bundled CLI binary present but does not include host platform; falling back to remote install",
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
func (h *handler) detectAndLogMirror(archBase string) string {
	path := workspace.DetectFilesystemMirror(archBase)
	platforms := workspace.MirrorPlatforms(archBase)
	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	switch {
	case path != "":
		h.l.Info("terraform: build-vendored provider mirror detected, using airgap resolution",
			zap.String("arch_base", archBase),
			zap.String("mirror_path", path),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	case len(platforms) > 0:
		h.l.Warn("terraform: provider mirror present but does not include host platform; falling back to direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	default:
		h.l.Info("terraform: no provider mirror in artifact, using direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
		)
	}
	return path
}

// getWorkspace returns a valid workspace for working with this plugin
func (h *handler) getWorkspace() (workspace.Workspace, error) {
	plan := h.state.plan
	sandboxCfg := h.state.sandboxCfg

	archDir := h.state.workspace.Source().AbsPath()
	if plan.LocalArchive != nil {
		archDir = plan.LocalArchive.Path
	}

	arch, err := dirarchive.New(h.v,
		dirarchive.WithPath(archDir),
		dirarchive.WithIgnoreDotTerraformDir(),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithAddBackendFile("http"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(h.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: h.cfg.RunnerAPIURL,
		WorkspaceID: h.state.plan.TerraformBackend.WorkspaceID,
		Token:       h.cfg.RunnerAPIToken,
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := h.buildBinary(archDir, sandboxCfg.TerraformVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(h.v,
		staticvars.WithFileVars(plan.Vars),
		staticvars.WithEnvVars(plan.EnvVars),
		staticvars.WithFiles(plan.VarsFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(h.v,
		authvars.WithAWSAuth(h.state.auth.AWSAuth),
		authvars.WithAzureAuth(h.state.auth.AzureAuth),
		authvars.WithGCPAuth(h.state.auth.GCPAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if plan.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(h.v,
			shell.WithRunAuth(h.state.auth.AWSAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		// Auto-detect the build-time provider mirror inside the unpacked
		// OCI artifact. When the build runner shipped one (gated by an
		// org feature flag server-side), this swaps in a .terraformrc
		// pointing at it; otherwise it's a no-op and terraform init
		// fetches providers from registry.terraform.io as before.
		workspace.WithFilesystemMirror(h.detectAndLogMirror(archDir)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

// getWorkspace returns a valid workspace for working with this plugin when we have a plan
// NOTE: these are kept distinct in case they continue to evolve separately and to make it easier to reason about
func (h *handler) getWorkspaceWithPlan(planBytes []byte) (workspace.Workspace, error) {
	plan := h.state.plan
	sandboxCfg := h.state.sandboxCfg

	archDir := h.state.workspace.Source().AbsPath()
	if plan.LocalArchive != nil {
		archDir = plan.LocalArchive.Path
	}

	arch, err := dirarchive.New(h.v,
		dirarchive.WithPath(archDir),
		dirarchive.WithIgnoreDotTerraformDir(),
		dirarchive.WithIgnoreTerraformStateFile(),
		dirarchive.WithAddBackendFile("http"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create local archive: %w", err)
	}

	back, err := httpbackend.New(h.v, httpbackend.WithNuonTerraformWorkspaceConfig(&httpbackend.NuonWorkspaceConfig{
		APIEndpoint: h.cfg.RunnerAPIURL,
		WorkspaceID: h.state.plan.TerraformBackend.WorkspaceID,
		Token:       h.cfg.RunnerAPIToken,
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := h.buildBinary(archDir, sandboxCfg.TerraformVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	vars, err := staticvars.New(h.v,
		staticvars.WithFileVars(plan.Vars),
		staticvars.WithEnvVars(plan.EnvVars),
		staticvars.WithFiles(plan.VarsFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create variable set: %w", err)
	}

	authVars, err := authvars.New(h.v,
		authvars.WithAWSAuth(h.state.auth.AWSAuth),
		authvars.WithAzureAuth(h.state.auth.AzureAuth),
		authvars.WithGCPAuth(h.state.auth.GCPAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create auth vars: %w", err)
	}

	var hooks hooks.Hooks
	if plan.Hooks == nil {
		hooks = noop.New()
	} else {
		hooks, err = shell.New(h.v,
			shell.WithRunAuth(h.state.auth.AWSAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := workspace.New(h.v,
		workspace.WithHooks(hooks),
		workspace.WithArchive(arch),
		workspace.WithBackend(back),
		workspace.WithBinary(bin),
		workspace.WithVariables(vars),
		workspace.WithVariables(authVars),
		// See getWorkspace for an explanation of the filesystem mirror.
		workspace.WithFilesystemMirror(h.detectAndLogMirror(archDir)),
		workspace.WithPlanBytes(planBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
