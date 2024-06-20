package shell

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace/output"
)

func (s *shell) setCredentials(ctx context.Context) error {
	envVars, err := credentials.FetchEnv(ctx, s.Auth)
	if err != nil {
		return fmt.Errorf("unable to fetch environment vars: %w", err)
	}

	s.EnvVars = generics.MergeMap(s.EnvVars, envVars)
	return nil
}

func (s *shell) execScript(ctx context.Context, filename string, log hclog.Logger) error {
	if err := s.setCredentials(ctx); err != nil {
		return fmt.Errorf("unable to set credentials for exec: %w", err)
	}

	out, err := output.New(s.v, output.WithLogger(log))
	if err != nil {
		return fmt.Errorf("unable to get output: %w", err)
	}

	outWriter, err := out.Writer()
	if err != nil {
		return fmt.Errorf("unable to create output writer: %w", err)
	}

	fp := filepath.Join(s.rootDir, filename)
	cmd, err := command.New(s.v,
		command.WithEnv(s.EnvVars),
		command.WithCmd(fp),
		command.WithCwd(s.rootDir),
		command.WithStdout(outWriter),
		command.WithStdout(outWriter),
		command.WithStdin(os.Stdin),
	)
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	err = cmd.Exec(ctx)
	if err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	return nil
}
