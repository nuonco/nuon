package opentofu

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkgopentofuremote "github.com/nuonco/nuon/pkg/opentofu/binary/remote"
	pkgopentofuworkspace "github.com/nuonco/nuon/pkg/opentofu/workspace"
	dirarchive "github.com/nuonco/nuon/pkg/terraform/archive/dir"
	httpbackend "github.com/nuonco/nuon/pkg/terraform/backend/http"
	"github.com/nuonco/nuon/pkg/terraform/binary"
	"github.com/nuonco/nuon/pkg/terraform/hooks"
	"github.com/nuonco/nuon/pkg/terraform/hooks/noop"
	"github.com/nuonco/nuon/pkg/terraform/hooks/shell"
	authvars "github.com/nuonco/nuon/pkg/terraform/variables/auth"
	staticvars "github.com/nuonco/nuon/pkg/terraform/variables/static"
	tfworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

// sandboxDefaultOpenTofuVersion is the OpenTofu CLI version used when the
// sandbox config does not pin one. The leading "v" is stripped by
// pkgopentofuremote.WithVersion.
const sandboxDefaultOpenTofuVersion = "v1.13.9"

// sandboxRequestedVersion resolves the OpenTofu CLI version for a sandbox
// run: the pinned open_tofu_version from the sandbox config when set, else
// the package default (mirroring how the terraform sandbox reads
// sandboxCfg.TerraformVersion).
func (h *handler) sandboxRequestedVersion() string {
	if h.state.sandboxCfg != nil && h.state.sandboxCfg.OpenTofuVersion != "" {
		return h.state.sandboxCfg.OpenTofuVersion
	}
	return sandboxDefaultOpenTofuVersion
}

// buildBinary always installs the OpenTofu CLI via tofudl. Unlike terraform,
// v1 has no build-vendored-binary convention for opentofu artifacts yet, so
// there is no bundled-binary detection to fall back around — the remote
// download is the only path (same as the opentofu deploy handler).
func (h *handler) buildBinary(version string) (binary.Binary, error) {
	return pkgopentofuremote.New(h.v, pkgopentofuremote.WithVersion(version))
}

// detectAndLogMirror runs the shared filesystem-mirror detection over the
// unpacked artifact and emits a single Info log describing which
// provider-resolution path the install runner will take. The detection is
// purely filesystem-driven and tool-agnostic: the build runner vendors
// providers into the same conventional directory for both engines, so
// tfworkspace.DetectFilesystemMirror finds an opentofu mirror too. The
// returned value is suitable to pass to pkgopentofuworkspace.WithFilesystemMirror.
func (h *handler) detectAndLogMirror(archBase string) string {
	path := tfworkspace.DetectFilesystemMirror(archBase)
	platforms := tfworkspace.MirrorPlatforms(archBase)
	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	switch {
	case path != "":
		h.l.Info("opentofu: build-vendored provider mirror detected, using airgap resolution",
			zap.String("arch_base", archBase),
			zap.String("mirror_path", path),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	case len(platforms) > 0:
		h.l.Warn("opentofu: provider mirror present but does not include host platform; falling back to direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
			zap.Strings("mirror_platforms", platforms),
		)
	default:
		h.l.Info("opentofu: no provider mirror in artifact, using direct registry resolution",
			zap.String("arch_base", archBase),
			zap.String("host_platform", hostPlatform),
		)
	}
	return path
}

// getWorkspace returns a valid workspace for working with this plugin
func (h *handler) getWorkspace() (tfworkspace.Workspace, error) {
	plan := h.state.plan

	var archDir string
	if h.state.ociArch != nil {
		archDir = h.state.ociArch.BasePath()
	} else {
		archDir = h.state.workspace.Source().AbsPath()
	}
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
		Token:       h.cfg.APIToken(),
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := h.buildBinary(h.sandboxRequestedVersion())
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

	var hooksv hooks.Hooks
	if plan.Hooks == nil {
		hooksv = noop.New()
	} else {
		hooksv, err = shell.New(h.v,
			shell.WithRunAuth(h.state.auth.AWSAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := pkgopentofuworkspace.New(h.v,
		pkgopentofuworkspace.WithHooks(hooksv),
		pkgopentofuworkspace.WithArchive(arch),
		pkgopentofuworkspace.WithBackend(back),
		pkgopentofuworkspace.WithBinary(bin),
		pkgopentofuworkspace.WithVariables(vars),
		pkgopentofuworkspace.WithVariables(authVars),
		// Auto-detect the build-time provider mirror inside the unpacked
		// OCI artifact, mirroring the terraform sandbox. Empty path = no-op.
		pkgopentofuworkspace.WithFilesystemMirror(h.detectAndLogMirror(archDir)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}

// getWorkspaceWithPlan returns a valid workspace for working with this plugin when we have a plan
// NOTE: these are kept distinct in case they continue to evolve separately and to make it easier to reason about
func (h *handler) getWorkspaceWithPlan(planBytes []byte) (tfworkspace.Workspace, error) {
	plan := h.state.plan

	var archDir string
	if h.state.ociArch != nil {
		archDir = h.state.ociArch.BasePath()
	} else {
		archDir = h.state.workspace.Source().AbsPath()
	}
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
		Token:       h.cfg.APIToken(),
		JobID:       h.state.jobID,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get http backend")
	}

	bin, err := h.buildBinary(h.sandboxRequestedVersion())
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

	var hooksv hooks.Hooks
	if plan.Hooks == nil {
		hooksv = noop.New()
	} else {
		hooksv, err = shell.New(h.v,
			shell.WithRunAuth(h.state.auth.AWSAuth),
			shell.WithEnvVars(plan.Hooks.EnvVars),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get hooks: %w", err)
		}
	}

	wkspace, err := pkgopentofuworkspace.New(h.v,
		pkgopentofuworkspace.WithHooks(hooksv),
		pkgopentofuworkspace.WithArchive(arch),
		pkgopentofuworkspace.WithBackend(back),
		pkgopentofuworkspace.WithBinary(bin),
		pkgopentofuworkspace.WithVariables(vars),
		pkgopentofuworkspace.WithVariables(authVars),
		// See getWorkspace for an explanation of the filesystem mirror.
		pkgopentofuworkspace.WithFilesystemMirror(h.detectAndLogMirror(archDir)),
		pkgopentofuworkspace.WithPlanBytes(planBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	return wkspace, nil
}
