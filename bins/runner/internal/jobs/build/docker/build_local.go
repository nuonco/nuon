package docker

import (
	"context"
	"fmt"
	"os/exec"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

func (b *handler) dockerPath() (string, error) {
	bins := []string{
		"docker",
		"podman",
	}
	for _, bin := range bins {
		path, err := exec.LookPath(bin)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no valid podman or docker path found")
}

func (b *handler) buildLocal(
	ctx context.Context,
	log *zap.Logger,
	dockerfilePath string,
	contextDir string,
	buildArgs map[string]*string,
	localRef string,
) error {
	dockerPath, err := b.dockerPath()
	if err != nil {
		return err
	}

	args := []string{
		"build",
		".",
		"-f", dockerfilePath,
		"--tag", localRef,
	}
	if b.state.cfg.Target != "" {
		args = append(args, "--target", b.state.cfg.Target)
	}
	// If we have build args we append each
	for k, v := range buildArgs {
		// v should always not be nil but guard just in case to avoid a panic
		if v != nil {
			args = append(args, "--build-arg", k+"="+*v)
		}
	}

	lf := zapwriter.New(log, zapcore.InfoLevel, "kaniko-build")
	cmd, err := command.New(b.v,
		command.WithCmd(dockerPath),
		command.WithCwd(contextDir),
		command.WithArgs(args),
		command.WithEnv(map[string]string{}),
		command.WithStdout(lf),
		command.WithStderr(lf),
	)
	if err != nil {
		return fmt.Errorf("unable to create build command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to build: %w", err)
	}

	return nil
}
