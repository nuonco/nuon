package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/actions/supervisor"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/zapwriter"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const containerWorkspaceMount = "/nuon/work"

// actionCPUShares is half docker's default weight of 1024, so an action can use
// idle CPU but loses to the install container when the VM is saturated.
const actionCPUShares = 512

// execCommandInContainer runs an image-backed action step inside its container.
// The workspace is bind-mounted so the supervisor writes outputs to a file the
// runner reads back on the host after the container exits.
func (h *handler) execCommandInContainer(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, envVars map[string]string) error {
	if h.launcher == nil {
		return errors.New("image-backed action received by a runner without a container launcher")
	}

	scriptHostPath, err := h.prepareInlineContentsCommand(ctx, l, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to prepare inline command")
	}

	root := h.state.workspace.Root()

	// write the supervisor shell script into the workspace (already bind-mounted)
	supervisorHostPath, err := supervisor.Write(root)
	if err != nil {
		return errors.Wrap(err, "unable to write actions supervisor")
	}

	mapPath := func(hostPath string) string {
		rel, relErr := filepath.Rel(root, hostPath)
		if relErr != nil {
			return hostPath
		}
		return filepath.Join(containerWorkspaceMount, rel)
	}

	builtInEnv, err := h.getContainerBuiltInEnv(ctx, cfg, mapPath)
	if err != nil {
		return errors.Wrap(err, "unable to get container env")
	}

	// same env layering as the host path, minus any inherited host environment.
	env := map[string]string{"COLUMNS": "500"}
	env = generics.MergeMap(env, h.state.plan.BuiltinEnvVars)
	env = generics.MergeMap(env, builtInEnv)
	env = generics.MergeMap(env, h.state.run.RunEnvVars)
	env = generics.MergeMap(env, envVars)
	env = generics.MergeMap(env, h.state.plan.OverrideEnvVars)

	image, err := h.actionImageRef()
	if err != nil {
		return err
	}

	outL := l.With(zap.String("nuon.command_output", "true"))
	lOut := zapwriter.NewWithOpts(outL, zapwriter.WithLogLevel(zapcore.InfoLevel), zapwriter.WithLineBuffering())
	lErr := zapwriter.NewWithOpts(outL, zapwriter.WithLogLevel(zapcore.ErrorLevel), zapwriter.WithLineBuffering())

	spec := launcher.RunSpec{
		Image:         image,
		ContainerName: fmt.Sprintf("nuon-action-%s-%d-%s", h.state.run.ID, cfg.Idx, randContainerSuffix()),
		Mounts: []launcher.Mount{
			{HostPath: root, ContainerPath: containerWorkspaceMount},
		},
		// run the supervisor via the image's own /bin/sh so it works in any
		// base image (musl/glibc/any arch) — no mounted binary to exec.
		Command: []string{
			"/bin/sh", mapPath(supervisorHostPath),
			"--script", mapPath(scriptHostPath),
			"--workdir", containerWorkspaceMount,
		},
		Env: env,
		Labels: map[string]string{
			"nuon.install_id": h.state.plan.InstallID,
			"nuon.run_id":     h.state.run.ID,
		},
		// No hard CPU quota: an action should be able to use otherwise idle VM
		// CPU. mng and the install container get relative priority through
		// CPUShares instead, so they win under contention without capping
		// actions to a single core.
		Memory:    "2g",
		CPUShares: actionCPUShares,
		PidsLimit: 512,
		Stdout:    lOut,
		Stderr:    lErr,
	}

	opCtx, end := op.Tool(ctx, "action", "image-command")
	runErr := h.launcher.Run(opCtx, spec)
	lOut.Flush()
	lErr.Flush()
	if runErr != nil {
		end(runErr)
		return fmt.Errorf("unable to run action container: %w", runErr)
	}
	end(nil)

	return nil
}

// prepareActionImage pulls the action's image once for the whole job and leases
// it against host image collection. Every step then reuses that one pull, and
// the image survives between steps instead of being re-pulled per step.
func (h *handler) prepareActionImage(ctx context.Context, l *zap.Logger, leaseID string) error {
	if h.launcher == nil {
		return errors.New("image-backed action received by a runner without a container launcher")
	}

	image, err := h.actionImageRef()
	if err != nil {
		return err
	}

	username, password, err := h.actionImagePullAuth(ctx)
	if err != nil {
		return err
	}

	l.Info("preparing image-backed action image", zap.String("action.image", image))

	return h.launcher.Prepare(ctx, launcher.PrepareSpec{
		Image:        image,
		PullUsername: username,
		PullPassword: password,
		LeaseID:      leaseID,
		// pull progress: INFO, untagged (not the action's command output)
		PullLog: zapwriter.New(l, zapcore.InfoLevel, ""),
	})
}

// releaseActionImage drops the job's lease. The image stays in the host cache so
// the next run of the same action skips the pull.
func (h *handler) releaseActionImage(leaseID string) {
	if h.launcher == nil {
		return
	}
	h.launcher.Release(leaseID)
}

// actionImageRef resolves the image ref the launcher runs. Production requires
// the digest-pinned ref that ctl-api resolved, so a step can only ever run the
// exact manifest Nuon resolved. There is deliberately no mutable-tag fallback:
// without a digest we fail rather than run whatever the tag happens to point at
// now. The dev-only real-docker path pulls the app-authored source image
// directly.
func (h *handler) actionImageRef() (string, error) {
	plan := h.state.plan

	if h.pullSourceImageDirectly() {
		return plan.SourceImage, nil
	}

	if plan.ImageDigestRef == "" {
		return "", errors.New("image-backed action plan is not pinned to an image digest")
	}

	return plan.ImageDigestRef, nil
}

// actionImagePullAuth resolves registry credentials for the pull. It goes
// through the registry package rather than reading OCIAuth off the plan: a
// mirrored image lands in the org registry, which uses static credentials, but
// a component-backed image is pulled straight from the install's own ECR, ACR,
// or GAR, which mint a token per pull.
func (h *handler) actionImagePullAuth(ctx context.Context) (username, password string, err error) {
	plan := h.state.plan

	if h.pullSourceImageDirectly() {
		return "", "", nil
	}

	if plan.ImageRegistry == nil {
		return "", "", errors.New("image-backed action plan has no image registry")
	}

	accessInfo, err := oci.FetchAccessInfo(ctx, plan.ImageRegistry)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get registry credentials for the action image")
	}
	if accessInfo.Auth == nil {
		return "", "", nil
	}

	return accessInfo.Auth.Username, accessInfo.Auth.Password, nil
}

// randContainerSuffix returns a short random suffix so concurrent executions of
// the same run+step get distinct container names. Without it, two overlapping
// attempts share a deterministic name and each attempt's docker rm -f would
// kill the other's live container.
func randContainerSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "x"
	}
	return hex.EncodeToString(b)
}

// pullSourceImageDirectly reports whether the dev-only real-docker path is
// active, in which case the launcher pulls the app-authored source image
// directly (ctl-api can't know a local registry address) instead of the
// mirror.
//
// It deliberately does not apply to an image that already resolved to the
// install's own registry, which is the case when the plan pins the source ref
// itself. That ref is never publicly pullable, so taking the shortcut would
// skip credential resolution and 401 rather than fall back to anything.
func (h *handler) pullSourceImageDirectly() bool {
	plan := h.state.plan
	if plan.ImageDigestRef != "" && plan.ImageDigestRef == plan.SourceImage {
		return false
	}

	return os.Getenv("NUON_DEV_REAL_IMAGE_ACTIONS") == "true" &&
		strings.EqualFold(os.Getenv("ENV"), "development")
}
