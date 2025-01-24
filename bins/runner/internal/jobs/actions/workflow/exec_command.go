package workflow

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/git"
	"github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon-runner-go/models"
)

func (h *handler) execCommand(ctx context.Context, l *zap.Logger, cfg *models.AppActionWorkflowStepConfig, src *git.Source) error {
	builtInEnv, err := h.getBuiltInEnv(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get execution env")
	}
	for k, v := range builtInEnv {
		l.Debug(fmt.Sprintf("setting default env-var %s", k), zap.String("value", v))
	}

	cmd, args, err := h.parseCommand(ctx, l, cfg, src)
	if err != nil {
		return errors.Wrap(err, "unable to parse command")
	}

	lOut := zapwriter.New(l, zapcore.InfoLevel, "")
	lErr := zapwriter.New(l, zapcore.ErrorLevel, "")
	dirName := git.Dir(src)
	cwd := h.state.workspace.AbsPath(dirName)

	cmdP, err := command.New(h.v,
		command.WithEnv(builtInEnv),
		command.WithCwd(cwd),
		command.WithCmd(cmd),
		command.WithArgs(args[0:]),
		command.WithCmd(cmd),
		command.WithInheritedEnv(),
		command.WithEnv(builtInEnv),
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
