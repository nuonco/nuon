package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/powertoolsdev/mono/pkg/command"
	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"
)

const (
	nuonCommandName string = "nuon"
)

type ExecTestScriptRequest struct {
	Path string            `validate:"required"`
	Env  map[string]string `validate:"required"`

	TFOutputsPath string               `validate:"required"`
	TFOutputs     *TerraformRunOutputs `validate:"required"`

	InstallCLI bool `validate:"required"`
}

type ExecTestScriptResponse struct{}

func (a *Activities) ExecTestScript(ctx context.Context, req *ExecTestScriptRequest) (*ExecTestScriptResponse, error) {
	l := activity.GetLogger(ctx)
	l.Info("executing test", zap.String("path", req.Path))
	if req.InstallCLI {
		if err := a.installCLI(ctx); err != nil {
			return nil, fmt.Errorf("unable to install cli: %w", err)
		}
	}

	if err := a.writeTFOutputs(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to write terraform outputs: %w", err)
	}

	err := a.execTestScript(ctx, req)
	if err != nil {
		l.Info("error executing test", zap.String("path", req.Path), zap.Error(err))
		return nil, fmt.Errorf("unable to execute cli command: %w", err)
	}

	return &ExecTestScriptResponse{}, nil
}

func (a *Activities) installCLI(ctx context.Context) error {
	cmd, err := command.New(a.v,
		command.WithInheritedEnv(),
		command.WithCmd(a.cfg.InstallScriptPath),
		command.WithArgs([]string{}),
		command.WithStdout(os.Stdout),
		command.WithStdout(os.Stderr),
		command.WithStdin(nil),
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

func (a *Activities) writeTFOutputs(ctx context.Context, req *ExecTestScriptRequest) error {
	byts, err := json.Marshal(req.TFOutputs)
	if err != nil {
		return fmt.Errorf("unable to convert outputs to json: %w", err)
	}

	if err := os.WriteFile(req.TFOutputsPath, byts, 0644); err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}

	return nil
}

func (a *Activities) execTestScript(ctx context.Context, req *ExecTestScriptRequest) error {
	cmd, err := command.New(a.v,
		command.WithInheritedEnv(),
		command.WithEnv(req.Env),
		command.WithCmd(req.Path),
		command.WithArgs([]string{}),
		command.WithStdout(os.Stdout),
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
