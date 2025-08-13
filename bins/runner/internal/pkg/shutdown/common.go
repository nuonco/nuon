package shutdown

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-playground/validator/v10"
	pkgcommand "github.com/powertoolsdev/mono/pkg/command"
	"github.com/powertoolsdev/mono/pkg/zapwriter"
)

// NOTE: cannibalized from github.com/plackemacher/system-shutdown
func runCommand(ctx context.Context, l *zap.Logger, v *validator.Validate, command string, args ...string) (err error) {
	lf := zapwriter.New(l, zapcore.InfoLevel, "shutdown")

	cmd, err := pkgcommand.New(v,
		pkgcommand.WithCmd(command),
		pkgcommand.WithArgs(args),
		pkgcommand.WithStdout(lf),
		pkgcommand.WithStderr(lf),
	)

	if err != nil {
		return fmt.Errorf("unable to create shutdown: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to shutdown: %w", err)
	}
	return err
}
