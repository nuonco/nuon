package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/git"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/pkg/command"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/zapwriter"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (h *handler) execCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *plantypes.GitSource, envVars map[string]string) error {
	defaultEnvVars := map[string]string{
		"COLUMNS": "500",
	}

	builtInEnv, err := h.getBuiltInEnv(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to get execution env")
	}

	for k, v := range h.state.plan.BuiltinEnvVars {
		l.Debug(fmt.Sprintf("setting built-in env-var %s", k), zap.String("value", v))
	}
	for k, v := range builtInEnv {
		l.Debug(fmt.Sprintf("setting default env-var %s", k), zap.String("value", v))
	}
	for k, v := range envVars {
		l.Debug(fmt.Sprintf("setting env-var %s", k), zap.String("value", v))
	}
	for k, v := range h.state.run.RunEnvVars {
		l.Debug(fmt.Sprintf("setting extra env-var %s", k), zap.String("value", v))
	}
	for k, v := range h.state.plan.OverrideEnvVars {
		l.Debug(fmt.Sprintf("setting override env-var %s", k), zap.String("value", v))
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

	// Tag the user's script stdout/stderr so the UI can show just the command
	// output and hide the runner's own job-lifecycle logs (which share the
	// same oteljob scope and Info severity).
	outL := l.With(zap.String("nuon.command_output", "true"))
	lOut := zapwriter.New(outL, zapcore.InfoLevel, "")
	lErr := zapwriter.New(outL, zapcore.ErrorLevel, "")

	dirName := git.Dir(src)
	cwd := h.state.workspace.AbsPath(dirName)

	cmdP, err := command.New(h.v,
		command.WithCwd(cwd),
		command.WithCmd(cmd),
		command.WithArgs(args[0:]),
		command.WithCmd(cmd),
		command.WithInheritedEnv(),
		command.WithEnv(h.state.plan.BuiltinEnvVars),
		command.WithEnv(builtInEnv),
		command.WithEnv(h.state.run.RunEnvVars),
		command.WithEnv(envVars),
		command.WithEnv(h.state.plan.OverrideEnvVars),
		command.WithEnv(defaultEnvVars),
		command.WithArgs(args),
		command.WithStdout(lOut),
		command.WithStderr(lErr),
	)
	if err != nil {
		l.Error("error creating command", zap.Error(err))
		return fmt.Errorf("unable to create command: %w", err)
	}

	opCtx, end := op.Tool(ctx, "action", "command")
	if err := cmdP.Exec(opCtx); err != nil {
		end(err)
		l.Error("error executing command "+err.Error(), zap.Error(err))
		return fmt.Errorf("unable to execute command: %w", err)
	}
	end(nil)

	return nil
}
