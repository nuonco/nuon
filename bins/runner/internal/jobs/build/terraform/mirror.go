package terraform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/terraform/workspace"
)

const (
	// defaultMirrorTerraformVersion is the version of the terraform CLI
	// the build runner falls back to installing when the build plan does
	// not specify one. Kept aligned with the default in the
	// TerraformModuleComponentConfig model so most components hit the
	// same install. Stored without a leading "v" to match the
	// un-prefixed format of TerraformModuleComponentConfig.Version.
	defaultMirrorTerraformVersion = "1.7.5"
)

// resolveMirrorPlatforms returns the `<os>_<arch>` platform set the build
// runner should vendor providers for. By default it returns the runner's
// own platform — an org's runners are homogeneous in practice (one node
// pool / instance family), so the install runner that consumes the artifact
// runs on the same arch as the build runner that produced it. Vendoring
// just one platform keeps the artifact small (~50–150MB per provider, not
// multiples) and works transparently for both production (linux_amd64 /
// linux_arm64) and local dev (darwin_arm64 on a mac running `nctl
// run-local`).
//
// The env var TERRAFORM_MIRROR_PLATFORMS overrides the default for the
// rare case where the build runner needs to ship more platforms than its
// own (heterogeneous orgs, cross-arch testing). Plumbed via
// internal.Config.TerraformMirrorPlatforms.
//
// Heterogeneous orgs that hit the default fall through to
// DetectFilesystemMirror's platform-mismatch branch on the install side —
// graceful warn + direct registry resolution.
func (h *handler) resolveMirrorPlatforms() []string {
	if h.cfg != nil && len(h.cfg.TerraformMirrorPlatforms) > 0 {
		return h.cfg.TerraformMirrorPlatforms
	}
	return []string{runtime.GOOS + "_" + runtime.GOARCH}
}

// scrubbedTFEnvVars are environment variables we strip from the build
// runner's environment before invoking the terraform CLI. They can otherwise
// silently redirect provider/module resolution to a host-side config or
// cache the build runner doesn't manage.
var scrubbedTFEnvVars = []string{
	"TF_CLI_CONFIG_FILE",
	"TF_PLUGIN_CACHE_DIR",
}

// generateProviderMirror installs a terraform CLI at the version requested
// by the build plan and produces an offline, install-runner-ready bundle in
// the source tree. The full sequence is:
//
//  1. `terraform get` — vendor remote modules into `<srcDir>/.terraform/modules/`
//     so `module {}` blocks resolve without registry/GitHub access at install
//     time. Run first because `providers mirror` resolves provider
//     requirements transitively through the module graph.
//
//  2. `terraform providers lock -platform=...` — create or augment
//     `<srcDir>/.terraform.lock.hcl` so it carries hashes for every install
//     runner platform we ship to. Most developers run `terraform init` on
//     macOS, which produces a lockfile with darwin hashes only; without this
//     step the install runner's `terraform init` would fail with "the cached
//     package for X does not match any of the checksums recorded in the
//     dependency lock file" on linux. The command honors existing version
//     pins and only adds the missing platform entries — it never bumps a
//     version that's already pinned.
//
//  3. `terraform providers mirror -platform=...` — vendor the locked provider
//     versions into `<srcDir>/<DefaultFilesystemMirrorDir>` for every platform
//     in mirrorPlatforms. With (2) above, the lockfile and the mirror always
//     agree on versions and hashes, so `terraform init` on the install runner
//     succeeds offline.
//
// The walker that packs the OCI artifact picks the modules tree, the
// lockfile, and the mirror tree up automatically because they all sit
// alongside the source.
//
// We invoke the CLI via os/exec rather than tfexec because tfexec v0.23
// does not expose the `providers mirror` subcommand.
func (h *handler) generateProviderMirror(ctx context.Context, srcDir string) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	tfVersion := defaultMirrorTerraformVersion
	if h.state.cfg != nil && h.state.cfg.TerraformVersion != "" {
		tfVersion = h.state.cfg.TerraformVersion
	}

	platforms := h.resolveMirrorPlatforms()

	l.Info("vendoring terraform providers and modules into the build artifact",
		zap.String("terraform_version", tfVersion),
		zap.String("src_dir", srcDir),
		zap.Strings("platforms", platforms),
	)

	execPath, cleanup, err := installTerraform(ctx, l, tfVersion)
	if err != nil {
		return fmt.Errorf("unable to install terraform binary for mirror: %w", err)
	}
	defer cleanup()

	// Vendor remote modules first. `terraform providers mirror` resolves
	// provider requirements transitively through the module graph, so the
	// modules need to be on disk before we mirror providers.
	if err := vendorModules(ctx, l, execPath, srcDir); err != nil {
		return fmt.Errorf("unable to vendor terraform modules: %w", err)
	}

	// Update the lockfile to carry hashes for every platform we mirror.
	// Done before `providers mirror` so the mirror downloads honor the
	// versions the lockfile pins.
	if err := lockProviders(ctx, l, execPath, srcDir, platforms); err != nil {
		return fmt.Errorf("unable to update terraform lockfile: %w", err)
	}

	mirrorDir := filepath.Join(srcDir, workspace.DefaultFilesystemMirrorDir)
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		return fmt.Errorf("unable to create mirror dir: %w", err)
	}

	args := []string{"providers", "mirror"}
	for _, p := range platforms {
		args = append(args, "-platform="+p)
	}
	args = append(args, mirrorDir)

	l.Info("running terraform providers mirror", zap.String("mirror_dir", mirrorDir))

	if err := runTerraform(ctx, l, execPath, srcDir, args...); err != nil {
		return fmt.Errorf("terraform providers mirror failed: %w", err)
	}

	// Now that the provider mirror is in place, also vendor the terraform
	// CLI binary itself so install runners can run fully airgapped (no
	// fetch from releases.hashicorp.com for `terraform_<ver>_<plat>.zip`
	// either). Reuses the host-platform binary we already installed above
	// to avoid a redundant download in the modal single-platform case.
	if err := vendorTerraformBinary(ctx, l, execPath, srcDir, tfVersion, platforms); err != nil {
		return fmt.Errorf("unable to vendor terraform binary: %w", err)
	}

	return nil
}

// vendorTerraformBinary copies the terraform CLI binary itself into
// `<srcDir>/<workspace.DefaultBundledBinaryDir>/<host>/terraform` and
// writes a sibling `VERSION` sidecar recording tfVersion. The artifact
// packer picks the tree up alongside everything else (mirror, modules,
// lockfile).
//
// We only vendor the host platform's binary, even when `platforms`
// includes others for the provider mirror. Reasons:
//
//  1. hc-install's ExactVersion does not expose OS/Arch overrides, so
//     cross-platform binary vendoring would require a manual HTTP fetch
//     against releases.hashicorp.com — not free, and not justified by the
//     modal use case (homogeneous orgs).
//  2. The install side is graceful about platform-mismatch artifacts:
//     workspace.DetectBundledBinary returns "" when the host platform's
//     binary is absent and the runner falls through to its existing
//     remotebinary path. So a heterogeneous setup that vendors providers
//     across platforms still works — just without the binary airgap on
//     non-build platforms.
//
// If we ever need cross-platform binary vendoring, the natural extension
// is a manual fetch from
// `https://releases.hashicorp.com/terraform/<ver>/terraform_<ver>_<os>_<arch>.zip`,
// gated by a TERRAFORM_BINARY_PLATFORMS env var.
func vendorTerraformBinary(
	ctx context.Context,
	l *zap.Logger,
	hostExecPath, srcDir, tfVersion string,
	platforms []string,
) error {
	binDir := filepath.Join(srcDir, workspace.DefaultBundledBinaryDir)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("unable to create bundled binary dir: %w", err)
	}

	hostPlatform := runtime.GOOS + "_" + runtime.GOARCH

	skipped := make([]string, 0)
	for _, p := range platforms {
		if p != hostPlatform {
			skipped = append(skipped, p)
		}
	}
	if len(skipped) > 0 {
		// Surface the limitation in the build log so heterogeneous-org
		// operators understand why install runners on non-build
		// platforms still hit releases.hashicorp.com for the CLI even
		// though their providers are vendored.
		l.Info("skipping CLI binary vendoring for non-host platforms (provider mirror still covers them)",
			zap.String("host_platform", hostPlatform),
			zap.Strings("skipped_platforms", skipped),
		)
	}

	dst := filepath.Join(binDir, hostPlatform, bundledBinaryName)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("unable to create bundled binary platform dir: %w", err)
	}
	if err := copyExecutable(hostExecPath, dst); err != nil {
		return fmt.Errorf("unable to copy terraform binary for %s: %w", hostPlatform, err)
	}
	l.Info("vendored terraform CLI binary",
		zap.String("platform", hostPlatform),
		zap.String("version", tfVersion),
		zap.String("dst", dst),
	)

	// VERSION sidecar: install side uses this to detect terraform_version
	// drift between the build that produced this artifact and the install
	// plan that's about to consume it. Single line — keep it trivial to
	// read & compare.
	versionPath := filepath.Join(binDir, workspace.BundledBinaryVersionFile)
	if err := os.WriteFile(versionPath, []byte(tfVersion+"\n"), 0o644); err != nil {
		return fmt.Errorf("unable to write bundled binary VERSION sidecar: %w", err)
	}

	return nil
}

// bundledBinaryName mirrors the unexported constant of the same name in
// pkg/terraform/workspace. Duplicated here to avoid exporting an
// otherwise-internal filename.
const bundledBinaryName = "terraform"

// copyExecutable copies src to dst byte-for-byte and chmods dst 0755 so
// the install runner sees an executable bit (preserved by OCI packing).
// We intentionally chmod after copy rather than mirroring src's mode:
// hc-install already produces 0755, but if a future src ever doesn't,
// the bundled binary still has to be executable to be useful.
func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("open dst: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return os.Chmod(dst, 0o755)
}

// terraformLockFile is the conventional name of the dependency lockfile
// terraform writes alongside a configuration. We read it before/after
// `terraform providers lock` so we can log clear "your lockfile changed"
// signals into the build log.
const terraformLockFile = ".terraform.lock.hcl"

// lockProviders runs `terraform providers lock -platform=...` against
// srcDir to ensure `.terraform.lock.hcl` carries hashes for every platform
// the install runner might run on. The command is platform-additive: it
// keeps existing platform entries and any version pins already recorded.
//
// We snapshot the lockfile bytes before and after so the build log clearly
// surfaces (a) whether the developer had committed a lockfile, and (b)
// whether running lock changed the on-disk file. This is the primary
// signal a developer has that the build runner is touching their lockfile
// on their behalf — the install side will silently consume whatever we
// produce here.
func lockProviders(ctx context.Context, l *zap.Logger, execPath, srcDir string, platforms []string) error {
	lockPath := filepath.Join(srcDir, terraformLockFile)

	pre, preErr := os.ReadFile(lockPath)
	switch {
	case preErr == nil:
		l.Info("found committed terraform lockfile; will augment with cross-platform hashes",
			zap.Int("size_bytes", len(pre)),
		)
	case errors.Is(preErr, os.ErrNotExist):
		l.Info("no .terraform.lock.hcl committed; build runner will generate one. " +
			"commit the generated lockfile for reproducible builds.")
	default:
		return fmt.Errorf("unable to read existing lockfile %s: %w", lockPath, preErr)
	}

	args := []string{"providers", "lock"}
	for _, p := range platforms {
		args = append(args, "-platform="+p)
	}

	l.Info("running terraform providers lock")

	if err := runTerraform(ctx, l, execPath, srcDir, args...); err != nil {
		return fmt.Errorf("terraform providers lock failed: %w", err)
	}

	post, err := os.ReadFile(lockPath)
	if err != nil {
		// `providers lock` ran cleanly but no lockfile? Treat as a hard
		// error — the install runner relies on its presence.
		return fmt.Errorf("expected lockfile at %s after providers lock: %w", lockPath, err)
	}

	switch {
	case preErr != nil:
		l.Info("generated fresh .terraform.lock.hcl",
			zap.Int("size_bytes", len(post)),
		)
	case bytes.Equal(pre, post):
		l.Info("lockfile already covered all build platforms; no changes")
	default:
		// Drift case: this is the modal happy path for the feature
		// (developer ran `terraform init` on macOS, lockfile gained the
		// linux hashes we need for install runners). We log at Info —
		// not Warn — because Warn'ing on the designed-for case would
		// make every successful first build look like an anomaly. The
		// "tip: commit the updated file" message gives the developer
		// what they need to lock things down further if they want to.
		l.Info("lockfile updated by terraform providers lock with cross-platform hashes; "+
			"tip: commit the updated lockfile for fully reproducible builds",
			zap.Int("pre_size_bytes", len(pre)),
			zap.Int("post_size_bytes", len(post)),
		)
	}

	return nil
}

// vendorModules runs `terraform get` against srcDir so any remote modules
// referenced by `module {}` blocks are downloaded into
// `<srcDir>/.terraform/modules/`. The packer picks the tree up alongside
// the source files. The install runner then unpacks it next to the source
// so `terraform init` finds modules locally and skips the registry fetch.
//
// Note we intentionally use `terraform get` rather than
// `terraform init -backend=false`: `init` would also try to negotiate with
// the configured backend (and our shipped source has no concrete backend
// config until the install runner writes one). `terraform get` is module-
// only and runs against the bare source tree.
//
// Caveat: modules sourced from private repositories (e.g.
// `git::ssh://git@github.com/...`) require git/SSH credentials in the build
// runner's environment. The same constraint applied before this feature —
// the install runner needed those creds at deploy time — but is now visible
// at build time instead.
func vendorModules(ctx context.Context, l *zap.Logger, execPath, srcDir string) error {
	l.Info("running terraform get to vendor modules")

	if err := runTerraform(ctx, l, execPath, srcDir, "get"); err != nil {
		return fmt.Errorf("terraform get failed: %w", err)
	}

	return nil
}

// runTerraform invokes the terraform CLI with the given args, working
// directory, and a scrubbed environment. Stdout/Stderr are streamed line-by-
// line through zap.
func runTerraform(ctx context.Context, l *zap.Logger, execPath, srcDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, execPath, args...)
	cmd.Dir = srcDir
	cmd.Env = scrubbedEnv(os.Environ())
	cmd.Stdout = newZapWriter(l, false)
	cmd.Stderr = newZapWriter(l, true)
	return cmd.Run()
}

// scrubbedEnv returns env with any TF CLI variables that could redirect
// provider/module resolution stripped out. We want the build runner's
// terraform invocations to use only the in-workspace .terraformrc (when
// any) we control — never a host-side override.
func scrubbedEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		drop := false
		for _, k := range scrubbedTFEnvVars {
			if strings.HasPrefix(kv, k+"=") {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, kv)
		}
	}
	return out
}

// installTerraform installs a fixed-version terraform CLI into a temp
// directory and returns its exec path. The returned cleanup func removes
// the install directory.
//
// The install dir is prefixed with `tf-build-` to avoid the runner's
// `workspace.CleanupAll` reset hook (which wipes any /tmp dir starting with
// `workspace`, `run`, `plugin`, or `archive`).
func installTerraform(ctx context.Context, l *zap.Logger, ver string) (string, func(), error) {
	tfVersion, err := version.NewVersion(ver)
	if err != nil {
		return "", nil, fmt.Errorf("invalid terraform version %q: %w", ver, err)
	}

	installDir, err := os.MkdirTemp("", "tf-build-")
	if err != nil {
		return "", nil, fmt.Errorf("unable to create terraform install dir: %w", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(installDir); err != nil {
			l.Warn("failed to clean up terraform install dir",
				zap.String("dir", installDir), zap.Error(err))
		}
	}

	installer := &releases.ExactVersion{
		Product:    product.Terraform,
		Version:    tfVersion,
		InstallDir: installDir,
	}

	execPath, err := installer.Install(ctx)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("unable to install terraform %s: %w", ver, err)
	}

	return execPath, cleanup, nil
}

// zapWriter adapts a *zap.Logger to io.Writer so it can be wired up as the
// Stdout/Stderr of an exec.Cmd. We split incoming buffers on newlines and
// emit one zap entry per non-empty line; without that split, multi-line
// terraform output would be flattened into single log entries containing
// embedded "\n", and partial-line writes would emit half-lines as their
// own entries. Stderr lines log at Warn, stdout at Info.
type zapWriter struct {
	l        *zap.Logger
	isStderr bool
}

func newZapWriter(l *zap.Logger, isStderr bool) *zapWriter {
	return &zapWriter{l: l, isStderr: isStderr}
}

func (z *zapWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if line == "" {
			continue
		}
		if z.isStderr {
			z.l.Warn("terraform: " + line)
		} else {
			z.l.Info("terraform: " + line)
		}
	}
	return len(p), nil
}
