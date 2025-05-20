package docker

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

func (b *handler) pushLocal(
	ctx context.Context,
	log *zap.Logger,
	localRef string,
) error {
	dockerPath, err := b.dockerPath()
	if err != nil {
		return err
	}

	args := []string{
		"push",
		localRef,
		"--tls-verify=false",
	}

	lw := zapwriter.New(log, zapcore.InfoLevel, "push ")
	cmd, err := command.New(b.v,
		command.WithCmd(dockerPath),
		command.WithArgs(args),
		command.WithEnv(map[string]string{}),
		command.WithStdout(lw),
		command.WithStderr(lw),
	)
	if err != nil {
		return fmt.Errorf("unable to create push command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to push: %w", err)
	}

	return nil
}
