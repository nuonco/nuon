package opentofu

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	pkgopentofuremote "github.com/nuonco/nuon/pkg/opentofu/binary/remote"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/log"
	"github.com/nuonco/nuon/pkg/runner/op"
	tfworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

const (
	// defaultMirrorOpenTofuVersion is the version of the OpenTofu CLI the
	// build runner falls back to installing when the build plan does not
	// specify one. Stored without a leading "v" to match the un-prefixed
	// format the opentofu component config carries in its Version field.
	defaultMirrorOpenTofuVersion = "1.13.9"

	// tofuInstallTimeout bounds the tofu CLI download via tofudl. It mirrors
	// the terraform build handler's 10-minute install budget: hc-install's
	// (and tofudl's) defaults are not enough for a ~30MB archive when
	// several component builds download concurrently.
	tofuInstallTimeout = 10 * time.Minute

	// tofuLockFile is the conventional name of the dependency lockfile both
	// terraform and OpenTofu write alongside a configuration. We read it
	// before/after `tofu providers lock` so we can log clear "your lockfile
	// changed" signals into the build log.
	tofuLockFile = ".terraform.lock.hcl"
)

type MirrorConfig struct {
	OpenTofuVersion string
	MirrorPlatforms []string
}

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
// The shared TERRAFORM_MIRROR_PLATFORMS env var (plumbed via
// internal.Config.TerraformMirrorPlatforms) overrides the default for the
// rare case where the build runner needs to ship more platforms than its
// own (heterogeneous orgs, cross-arch testing). Despite the name the knob
// is engine-agnostic — the platform list is arch-based — and both the
// terraform and opentofu build handlers consume it.
//
// Heterogeneous orgs that hit the default fall through to
// DetectFilesystemMirror's platform-mismatch branch on the install side —
// graceful warn + direct registry resolution.
func resolveMirrorPlatforms(configured []string) []string {
	if len(configured) > 0 {
		return configured
	}
	return []string{runtime.GOOS + "_" + runtime.GOARCH}
}

// scrubbedTofuEnvVars are environment variables we strip from the build
// runner's environment before invoking the tofu CLI. They can otherwise
// silently redirect provider/module resolution to a host-side config or
// cache the build runner doesn't manage. OpenTofu reads the terraform-
// compatible TF_ names (TF_CLI_CONFIG_FILE is the only CLI-config knob —
// there is no TOFU_-prefixed equivalent), and TOFU_ variants are scrubbed
// defensively.
var scrubbedTofuEnvVars = []string{
	"TOFU_CLI_CONFIG_FILE",
	"TOFU_PLUGIN_CACHE_DIR",
	"TF_CLI_CONFIG_FILE",
	"TF_PLUGIN_CACHE_DIR",
}

// generateProviderMirror installs a tofu CLI at the version requested by the
// build plan and produces an offline, install-runner-ready bundle in the
// source tree. The full sequence is:
//
//  1. `tofu get` — vendor remote modules into `<srcDir>/.terraform/modules/`
//     so `module {}` blocks resolve without registry/GitHub access at install
//     time. Run first because `providers mirror` resolves provider
//     requirements transitively through the module graph.
//
//  2. `tofu providers lock -platform=...` — create or augment
//     `<srcDir>/.terraform.lock.hcl` so it carries hashes for every install
//     runner platform we ship to. Most developers run `tofu init` on
//     macOS, which produces a lockfile with darwin hashes only; without this
//     step the install runner's `tofu init` would fail with "the cached
//     package for X does not match any of the checksums recorded in the
//     dependency lock file" on linux. The command honors existing version
//     pins and only adds the missing platform entries — it never bumps a
//     version that's already pinned.
//
//  3. `tofu providers mirror -platform=...` — vendor the locked provider
//     versions into `<srcDir>/<tfworkspace.DefaultFilesystemMirrorDir>` for
//     every platform in mirrorPlatforms. With (2) above, the lockfile and
//     the mirror always agree on versions and hashes, so `tofu init` on the
//     install runner succeeds offline.
//
// The walker that packs the OCI artifact picks the modules tree, the
// lockfile, and the mirror tree up automatically because they all sit
// alongside the source.
//
// The mirror directory is the terraform workspace package's conventional
// name (".terraform-providers"): the layout is arch-based and tool-agnostic,
// and the install side detects it with the same shared
// tfworkspace.DetectFilesystemMirror helper for both engines.
//
// Unlike the terraform port, this does NOT vendor the tofu CLI binary into
// the artifact: OpenTofu has no build-vendored-binary convention on the
// install side yet (the deploy handler always installs via tofudl), so a
// vendored binary would only be dead weight.
//
// We invoke the CLI via os/exec rather than tfexec because tfexec cannot
// drive tofu (its version parser requires a `Terraform vX` banner) and does
// not expose the `providers mirror` subcommand.
func (h *handler) generateProviderMirror(ctx context.Context, srcDir string) error {
	mirrorCfg := MirrorConfig{}
	if h.state.cfg != nil {
		mirrorCfg.OpenTofuVersion = h.state.cfg.OpenTofuVersion
	}
	if h.cfg != nil {
		mirrorCfg.MirrorPlatforms = h.cfg.TerraformMirrorPlatforms
	}
	return GenerateProviderMirror(ctx, srcDir, mirrorCfg)
}

// GenerateProviderMirror runs tofu with host filesystem access, not sourcefs
// confinement.
func GenerateProviderMirror(ctx context.Context, srcDir string, cfg MirrorConfig) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	tofuVersion := defaultMirrorOpenTofuVersion
	if cfg.OpenTofuVersion != "" {
		tofuVersion = cfg.OpenTofuVersion
	}

	platforms := resolveMirrorPlatforms(cfg.MirrorPlatforms)

	l.Info("vendoring opentofu providers and modules into the build artifact",
		zap.String("opentofu_version", tofuVersion),
		zap.String("src_dir", srcDir),
		zap.Strings("platforms", platforms),
	)

	var cleanup func()
	var tofuPath string
	if err := op.Run(ctx, "opentofu", "install_cli", func(ctx context.Context) error {
		tofuPath, cleanup, err = installTofu(ctx, l, tofuVersion)
		return err
	}); err != nil {
		return fmt.Errorf("unable to install opentofu binary for mirror: %w", err)
	}
	defer cleanup()

	// Vendor remote modules first. `tofu providers mirror` resolves
	// provider requirements transitively through the module graph, so the
	// modules need to be on disk before we mirror providers.
	if err := op.Run(ctx, "opentofu", "vendor_modules", func(ctx context.Context) error {
		return runTofu(ctx, l, tofuPath, srcDir, "get", nil, "")
	}); err != nil {
		return fmt.Errorf("unable to vendor opentofu modules: %w", err)
	}

	// Update the lockfile to carry hashes for every platform we mirror.
	// Done before `providers mirror` so the mirror downloads honor the
	// versions the lockfile pins.
	if err := op.Run(ctx, "opentofu", "lock_providers", func(ctx context.Context) error {
		return lockProviders(ctx, l, tofuPath, srcDir, platforms)
	}); err != nil {
		return fmt.Errorf("unable to update opentofu lockfile: %w", err)
	}

	if err := op.Run(ctx, "opentofu", "mirror_providers", func(ctx context.Context) error {
		mirrorDir := filepath.Join(srcDir, tfworkspace.DefaultFilesystemMirrorDir)
		if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
			return fmt.Errorf("unable to create mirror dir: %w", err)
		}

		l.Info("running tofu providers mirror", zap.String("mirror_dir", mirrorDir))

		return runTofu(ctx, l, tofuPath, srcDir, "providers mirror", platformFlags(platforms), mirrorDir)
	}); err != nil {
		return err
	}

	return nil
}

// platformFlags builds the repeated `-platform=<os>_<arch>` flag list both
// `providers lock` and `providers mirror` accept.
func platformFlags(platforms []string) []string {
	flags := make([]string, 0, len(platforms))
	for _, p := range platforms {
		flags = append(flags, "-platform="+p)
	}
	return flags
}

// lockProviders runs `tofu providers lock -platform=...` against srcDir to
// ensure `.terraform.lock.hcl` carries hashes for every platform the install
// runner might run on. The command is platform-additive: it keeps existing
// platform entries and any version pins already recorded.
//
// We snapshot the lockfile bytes before and after so the build log clearly
// surfaces (a) whether the developer had committed a lockfile, and (b)
// whether running lock changed the on-disk file. This is the primary signal
// a developer has that the build runner is touching their lockfile on their
// behalf — the install side will silently consume whatever we produce here.
func lockProviders(ctx context.Context, l *zap.Logger, tofuPath, srcDir string, platforms []string) error {
	lockPath := filepath.Join(srcDir, tofuLockFile)

	pre, preErr := os.ReadFile(lockPath)
	switch {
	case preErr == nil:
		l.Info("found committed opentofu lockfile; will augment with cross-platform hashes",
			zap.Int("size_bytes", len(pre)),
		)
	case errors.Is(preErr, os.ErrNotExist):
		l.Info("no .terraform.lock.hcl committed; build runner will generate one. " +
			"commit the generated lockfile for reproducible builds.")
	default:
		return fmt.Errorf("unable to read existing lockfile %s: %w", lockPath, preErr)
	}

	l.Info("running tofu providers lock")

	if err := runTofu(ctx, l, tofuPath, srcDir, "providers lock", platformFlags(platforms), ""); err != nil {
		return fmt.Errorf("tofu providers lock failed: %w", err)
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
		// (developer ran `tofu init` on macOS, lockfile gained the
		// linux hashes we need for install runners). We log at Info —
		// not Warn — because Warn'ing on the designed-for case would
		// make every successful first build look like an anomaly. The
		// "tip: commit the updated file" message gives the developer
		// what they need to lock things down further if they want to.
		l.Info("lockfile updated by tofu providers lock with cross-platform hashes; "+
			"tip: commit the updated lockfile for fully reproducible builds",
			zap.Int("pre_size_bytes", len(pre)),
			zap.Int("post_size_bytes", len(post)),
		)
	}

	return nil
}

// installTofu installs a fixed-version tofu CLI into a temp directory via
// tofudl and returns its exec path. The returned cleanup func removes the
// install directory.
//
// The install dir is prefixed with `tofu-build-` to keep it separate from
// workspace directories (which are prefixed with `workspace-`).
//
// tofudl has no built-in install budget, so the install context is capped at
// tofuInstallTimeout explicitly — the same reason the terraform build
// handler overrides hc-install's 30s default.
func installTofu(ctx context.Context, l *zap.Logger, ver string) (string, func(), error) {
	installCtx, cancel := context.WithTimeout(ctx, tofuInstallTimeout)
	defer cancel()

	installDir, err := os.MkdirTemp("", "tofu-build-")
	if err != nil {
		return "", nil, fmt.Errorf("unable to create opentofu install dir: %w", err)
	}
	cleanup := func() {
		if err := os.RemoveAll(installDir); err != nil {
			l.Warn("failed to clean up opentofu install dir",
				zap.String("dir", installDir), zap.Error(err))
		}
	}

	_, endInstall := op.Start(ctx, "opentofu", "binary_install",
		attribute.String("opentofu.version", ver),
		attribute.String("install.dir", installDir),
	)
	var execPath string
	installer, err := pkgopentofuremote.New(validator.New(), pkgopentofuremote.WithVersion(ver))
	if err == nil {
		execPath, err = installer.Install(installCtx, log.NewHClog(l), installDir)
	}
	if err == nil {
		if _, serr := os.Stat(execPath); serr != nil {
			err = fmt.Errorf("opentofu binary missing after install: %w", serr)
		}
	}
	endInstall(err)
	if err != nil {
		cleanup()
		if ctx.Err() != nil {
			return "", nil, fmt.Errorf("unable to install opentofu %s: %w (context: %w)", ver, err, ctx.Err())
		}
		return "", nil, fmt.Errorf("unable to install opentofu %s (download may have timed out after %s): %w", ver, tofuInstallTimeout, err)
	}

	return execPath, cleanup, nil
}

// runTofu runs a single tofu subcommand against srcDir while capturing its
// stdout and stderr into in-memory buffers, then emits exactly one zap entry
// summarising the command:
//
//   - Info on success ("tofu <name> completed")
//   - Warn on failure ("tofu <name> failed", with the wrapped error)
//
// Captured output is attached as `stdout` / `stderr` zap fields after a
// one-shot ANSI escape strip, so tofu's colored diagnostic boxes (`╷ │ ╵`)
// read as plain text in the log. Empty streams are omitted.
//
// args are the subcommand's flags; extra (e.g. a mirror directory) is
// appended positionally.
//
// tofuPath is the installed tofu binary from installTofu — the build
// container has no tofu on PATH, so the installed path must be used.
func runTofu(ctx context.Context, l *zap.Logger, tofuPath, srcDir, name string, args []string, extra string) error {
	fullArgs := make([]string, 0, len(args)+2)
	fullArgs = append(fullArgs, strings.Fields(name)...)
	fullArgs = append(fullArgs, args...)
	if extra != "" {
		fullArgs = append(fullArgs, extra)
	}

	cmd := exec.CommandContext(ctx, tofuPath, fullArgs...)
	cmd.Dir = srcDir
	cmd.Env = scrubbedTofuEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	fields := make([]zap.Field, 0, 3)
	if s := strings.TrimRight(stripANSI(stdout.String()), "\n"); s != "" {
		fields = append(fields, zap.String("stdout", s))
	}
	if s := strings.TrimRight(stripANSI(stderr.String()), "\n"); s != "" {
		fields = append(fields, zap.String("stderr", s))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Warn("tofu "+name+" failed", fields...)
		return err
	}
	l.Info("tofu "+name+" completed", fields...)
	return nil
}

// ansiEscapeRe matches CSI SGR sequences ("\x1b[…m"), which tofu uses to
// color its diagnostic envelope and severity labels. Stripped once at the
// end of each command so log search is grep-friendly.
var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}

// scrubbedTofuEnv returns env (in os.Environ "K=V" form) with any tofu CLI
// variables that could redirect provider/module resolution stripped out. We
// want the build runner's tofu invocations to use only config we control —
// never a host-side override.
func scrubbedTofuEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			continue
		}
		k := kv[:i]
		drop := false
		for _, sk := range scrubbedTofuEnvVars {
			if k == sk {
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
