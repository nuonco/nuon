package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/powertoolsdev/mono/pkg/command"
)

const (
	nuonCommandName string = "nuon"
)

type ExecTestScriptRequest struct {
	Path string
	Env  map[string]string

	TFOutputsPath string
	TFOutputs     *TerraformRunOutputs

	InstallCLI bool
}

type ExecTestScriptResponse struct {
	Output       []byte
	StringOutput string
	JSONOutput   interface{}
}

func (a *Activities) ExecTestScript(ctx context.Context, req *ExecTestScriptRequest) (*ExecTestScriptResponse, error) {
	if req.InstallCLI {
		if err := a.installCLI(ctx); err != nil {
			return nil, fmt.Errorf("unable to install cli: %w", err)
		}
	}

	if err := a.writeTFOutputs(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to write terraform outputs: %w", err)
	}

	output, jsonOutput, err := a.execTestScript(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute cli command: %w", err)
	}

	return &ExecTestScriptResponse{
		StringOutput: string(output),
		Output:       output,
		JSONOutput:   jsonOutput,
	}, nil
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

func (a *Activities) execTestScript(ctx context.Context, req *ExecTestScriptRequest) ([]byte, interface{}, error) {
	cmd, err := command.New(a.v,
		command.WithInheritedEnv(),
		command.WithEnv(req.Env),
		command.WithCmd(req.Path),
		command.WithArgs([]string{}),
		command.WithStdout(nil),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create command: %w", err)
	}

	output, err := cmd.ExecWithOutput(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to execute command: %w", err)
	}

	var (
		mapOutput  map[string]interface{}
		listOutput []interface{}
	)
	if err := json.Unmarshal(output, &mapOutput); err == nil {
		return output, mapOutput, nil
	}
	if err := json.Unmarshal(output, &listOutput); err == nil {
		return output, listOutput, nil
	}

	return output, nil, nil
}
