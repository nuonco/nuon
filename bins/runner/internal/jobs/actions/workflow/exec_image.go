package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/launcher"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/zapwriter"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	containerWorkspaceMount = "/nuon/work"
	containerSupervisorPath = "/nuon/bin/runner"
)

// execCommandInContainer runs an image-backed action step inside its container.
// The workspace is bind-mounted so the supervisor writes outputs to a file the
// runner reads back on the host after the container exits.
func (h *handler) execCommandInContainer(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, envVars map[string]string) error {
	if h.launcher == nil {
		return errors.New("image-backed action received by a runner without a container launcher")
	}

	supervisorPath, err := h.supervisorBinaryPath()
	if err != nil {
		return errors.Wrap(err, "unable to resolve supervisor binary path")
	}

	scriptHostPath, err := h.prepareInlineContentsCommand(ctx, l, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to prepare inline command")
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

	// same env layering as the host path, minus any inherited host environment.
	env := map[string]string{"COLUMNS": "500"}
	env = generics.MergeMap(env, h.state.plan.BuiltinEnvVars)
	env = generics.MergeMap(env, builtInEnv)
	env = generics.MergeMap(env, h.state.run.RunEnvVars)
	env = generics.MergeMap(env, envVars)
	env = generics.MergeMap(env, h.state.plan.OverrideEnvVars)

	image, username, password := h.actionImageRef()

	outL := l.With(zap.String("nuon.command_output", "true"))
	lOut := zapwriter.New(outL, zapcore.InfoLevel, "")
	lErr := zapwriter.New(outL, zapcore.ErrorLevel, "")

	spec := launcher.RunSpec{
		Image:         image,
		PullUsername:  username,
		PullPassword:  password,
		ContainerName: fmt.Sprintf("nuon-action-%s-%d", h.state.run.ID, cfg.Idx),
		Mounts: []launcher.Mount{
			{HostPath: supervisorPath, ContainerPath: containerSupervisorPath, ReadOnly: true},
			{HostPath: root, ContainerPath: containerWorkspaceMount},
		},
		Command: []string{
			containerSupervisorPath, "actions-supervisor",
			"--script", mapPath(scriptHostPath),
			"--workdir", containerWorkspaceMount,
		},
		Env: env,
		Labels: map[string]string{
			"nuon.install_id": h.state.plan.InstallID,
			"nuon.run_id":     h.state.run.ID,
		},
		Memory:    "2g",
		CPUs:      "1",
		PidsLimit: 512,
		Stdout:    lOut,
		Stderr:    lErr,
	}

	opCtx, end := op.Tool(ctx, "action", "image-command")
	if err := h.launcher.Run(opCtx, spec); err != nil {
		end(err)
		return fmt.Errorf("unable to run action container: %w", err)
	}
	end(nil)

	return nil
}

// actionImageRef resolves the image ref and pull credentials for the launcher.
// Prod pulls the install-registry mirror; the dev real-docker path pulls the
// source image directly (no mirror, no creds).
func (h *handler) actionImageRef() (image, username, password string) {
	plan := h.state.plan

	if h.pullSourceImageDirectly() {
		return plan.SourceImage, "", ""
	}

	if plan.ImageRegistry != nil && plan.ImageTag != "" {
		image = fmt.Sprintf("%s/%s:%s",
			strings.TrimSuffix(plan.ImageRegistry.LoginServer, "/"),
			plan.ImageRegistry.Repository,
			plan.ImageTag,
		)
		if plan.ImageRegistry.OCIAuth != nil {
			username = plan.ImageRegistry.OCIAuth.Username
			password = plan.ImageRegistry.OCIAuth.Password
		}
		return image, username, password
	}

	return plan.SourceImage, "", ""
}

// supervisorBinaryPath returns the host path to the supervisor binary to mount
// into the container. In the dev real-docker path this must be an explicit
// linux cross-build (the darwin host binary can't run in a linux container).
func (h *handler) supervisorBinaryPath() (string, error) {
	if h.pullSourceImageDirectly() {
		if p := os.Getenv("NUON_DEV_SUPERVISOR_BIN"); p != "" {
			return p, nil
		}
	}
	return os.Executable()
}

// pullSourceImageDirectly reports whether the dev-only real-docker path is
// active, in which case the launcher pulls the app-authored source image
// directly (ctl-api can't know a local registry address) instead of the
// install-registry mirror.
func (h *handler) pullSourceImageDirectly() bool {
	return os.Getenv("NUON_DEV_REAL_IMAGE_ACTIONS") == "true" &&
		strings.EqualFold(os.Getenv("ENV"), "development")
}
