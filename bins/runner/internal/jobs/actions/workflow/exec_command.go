package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	"github.com/powertoolsdev/mono/pkg/command"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/zapwriter"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) execCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource, envVars map[string]string) error {
	builtInEnv, err := h.getBuiltInEnv(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get execution env")
	}

	for k, v := range h.state.plan.EnvVars {
		l.Debug(fmt.Sprintf("setting built-in env-var %s", k), zap.String("value", v))
	}
	for k, v := range builtInEnv {
		l.Debug(fmt.Sprintf("setting default env-var %s", k), zap.String("value", v))
	}
	for k, v := range h.state.run.RunEnvVars {
		l.Debug(fmt.Sprintf("setting extra env-var %s", k), zap.String("value", v))
	}
	for k, v := range envVars {
		l.Debug(fmt.Sprintf("setting env-var %s", k), zap.String("value", v))
	}

	var cmd string
	var args []string
	if cfg.InlineContents == "" {
		cmd, args, err = h.parseCommand(ctx, l, cfg, src)
		if err != nil {
			return errors.Wrap(err, "unable to parse command")
		}
	} else {
		cmd, err = h.prepareInlineContentsCommand(ctx, l, cfg)
		if err != nil {
			return errors.Wrap(err, "unable to create inline command")
		}
	}

	lOut := zapwriter.New(l, zapcore.InfoLevel, "")
	lErr := zapwriter.New(l, zapcore.ErrorLevel, "")

	dirName := git.Dir(src)
	cwd := h.state.workspace.AbsPath(dirName)

	cmdP, err := command.New(h.v,
		command.WithCwd(cwd),
		command.WithCmd(cmd),
		command.WithArgs(args[0:]),
		command.WithCmd(cmd),
		command.WithInheritedEnv(),
		command.WithEnv(h.state.plan.EnvVars),
		command.WithEnv(builtInEnv),
		command.WithEnv(h.state.run.RunEnvVars),
		command.WithEnv(envVars),
		command.WithArgs(args),
		command.WithStdout(lOut),
		command.WithStderr(lErr),
	)
	if err != nil {
		return fmt.Errorf("unable to build command: %w", err)
	}

	if err := cmdP.Exec(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	return nil
}
