package docker

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/powertoolsdev/mono/pkg/command"
)

func (b *handler) pushLocal(
	ctx context.Context,
	log hclog.Logger,
	localRef string,
) error {
	dockerPath, err := b.dockerPath()
	if err != nil {
		return err
	}

	args := []string{
		"push",
		localRef,
	}

	lw := log.StandardWriter(&hclog.StandardLoggerOptions{})
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
