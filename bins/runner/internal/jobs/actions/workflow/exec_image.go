package workflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/actions/supervisor"
	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/zapwriter"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const containerWorkspaceMount = "/nuon/work"

const actionCPUShares = 512

func (h *handler) execCommandInContainer(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource, envVars map[string]string) error {
	if h.launcher == nil {
		return errors.New("image-backed action received by a runner without a container launcher")
	}

	root := h.state.workspace.Root()

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

	command, workdir, err := h.containerCommand(ctx, l, cfg, src, mapPath)
	if err != nil {
		return errors.Wrap(err, "unable to prepare container command")
	}

	spec := launcher.RunSpec{
		Image:         image,
		ContainerName: fmt.Sprintf("nuon-action-%s-%d-%s", h.state.run.ID, cfg.Idx, randContainerSuffix()),
		Mounts: []launcher.Mount{
			{HostPath: root, ContainerPath: containerWorkspaceMount},
		},
		Command: command,
		Workdir: workdir,
		Env:     env,
		Labels: map[string]string{
			"nuon.install_id": h.state.plan.InstallID,
			"nuon.run_id":     h.state.run.ID,
		},
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

func (h *handler) containerCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource, mapPath func(string) string) ([]string, string, error) {
	if cfg.InlineContents == "" && (src == nil || src.URL == "") {
		if args := directContainerCommand(cfg.Command); len(args) > 0 {
			return args, mapPath(h.state.workspace.Root()), nil
		}
	}

	script, workdir, args, err := h.prepareContainerStep(ctx, l, cfg, src)
	if err != nil {
		return nil, "", err
	}
	supervisorPath, err := supervisor.Write(h.state.workspace.Root())
	if err != nil {
		return nil, "", err
	}
	command := []string{"/bin/sh", mapPath(supervisorPath), "--script", mapPath(script), "--workdir", mapPath(workdir)}
	if len(args) > 0 {
		command = append(command, "--")
		command = append(command, args...)
	}
	return command, "", nil
}

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
		PullLog:      zapwriter.New(l, zapcore.InfoLevel, ""),
	})
}

func (h *handler) releaseActionImage(leaseID string) {
	if h.launcher == nil {
		return
	}
	h.launcher.Release(leaseID)
}

// why: actionImageRef resolves the image ref the launcher runs. It is always the
// digest-pinned ref that ctl-api resolved, so a step can only ever run the
// exact manifest Nuon resolved. There is deliberately no mutable-tag fallback:
// without a digest we fail rather than run whatever the tag happens to point at
// now.
func (h *handler) actionImageRef() (string, error) {
	plan := h.state.plan

	if plan.ImageDigestRef == "" {
		return "", errors.New("image-backed action plan is not pinned to an image digest")
	}

	return plan.ImageDigestRef, nil
}

func (h *handler) actionImagePullAuth(ctx context.Context) (username, password string, err error) {
	plan := h.state.plan

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

func randContainerSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "x"
	}
	return hex.EncodeToString(b)
}
