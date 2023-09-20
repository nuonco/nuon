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

type CLICommandRequest struct {
	OrgID	string
	AppID	string
	Args	[]string
	Install bool
	Json	bool
}

type CLICommandResponse struct {
	Output	     []byte
	StringOutput string
	JsonOutput   interface{}
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

func (a *Activities) CLICommand(ctx context.Context, req *CLICommandRequest) (*CLICommandResponse, error) {
	if req.Install {
		if err := a.installCLI(ctx); err != nil {
			return nil, fmt.Errorf("unable to install cli: %w", err)
		}
	}

	output, jsonOut, err := a.execCLICommand(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to execute cli command: %w", err)
	}

	return &CLICommandResponse{
		JsonOutput:   jsonOut,
		StringOutput: string(output),
		Output:       output,
	}, nil
}

func (a *Activities) execCLICommand(ctx context.Context, req *CLICommandRequest) ([]byte, interface{}, error) {
	env := map[string]string{
		"NUON_API_URL":   a.cfg.APIURL,
		"NUON_API_TOKEN": a.cfg.APIToken,
		"NUON_ORG_ID":	  req.OrgID,
	}
	if req.AppID != "" {
		env["NUON_APP_ID"] = req.AppID
	}

	cmd, err := command.New(a.v,
		command.WithInheritedEnv(),
		command.WithEnv(env),
		command.WithCmd(nuonCommandName),
		command.WithArgs(req.Args),
		command.WithStdout(nil),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create command: %w", err)
	}

	output, err := cmd.ExecWithOutput(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to execute command: %w", err)
	}

	if !req.Json {
		return output, nil, nil
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

	return nil, nil, fmt.Errorf("unable to convert response to json")
}
