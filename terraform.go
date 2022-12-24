package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-exec/tfexec"
)

const (
	varsFilename  string = "nuon.tfvars.json"
	stateFilename string = "state.tf"
)

type printfer interface {
	Printf(format string, v ...interface{})
}

type terraformExecutor interface {
	// init client accepts workingDir, execPath
	initClient(string, string) error
	setLogger(printfer)
	setStderr(io.Writer)
	setStdout(io.Writer)
	setEnvVars(map[string]string) error
	initModule(context.Context) error
	planModule(context.Context) error
	applyModule(context.Context) error
	destroyModule(context.Context) error
	outputs(context.Context) (map[string]interface{}, error)
}

type localTerraformExecutor struct {
	tfClient  terraformClient
	outputter outputter
}

var _ terraformExecutor = (*localTerraformExecutor)(nil)

type terraformClient interface {
	Init(context.Context, ...tfexec.InitOption) error
	Apply(context.Context, ...tfexec.ApplyOption) error
	Destroy(context.Context, ...tfexec.DestroyOption) error
	Plan(context.Context, ...tfexec.PlanOption) (bool, error)

	SetStderr(io.Writer)
	SetStdout(io.Writer)
	SetEnv(map[string]string) error
}

var _ terraformClient = (*tfexec.Terraform)(nil)

// initClient accepts a workingDir, execPath and initializes terraform
func (t *localTerraformExecutor) initClient(workingDir, execPath string) error {
	tfClient, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return err
	}
	t.tfClient = tfClient
	t.outputter = tfClient
	return nil
}

func (t *localTerraformExecutor) setStdout(f io.Writer) {
	t.tfClient.SetStdout(f)
}

func (t *localTerraformExecutor) setStderr(f io.Writer) {
	t.tfClient.SetStderr(f)
}

func (t *localTerraformExecutor) setLogger(pf printfer) {
	tfexecClient, ok := t.tfClient.(*tfexec.Terraform)
	if ok {
		tfexecClient.SetLogger(pf)
	}
}

func (t *localTerraformExecutor) setEnvVars(vars map[string]string) error {
	envVars := getEnv()
	for k, v := range vars {
		envVars[k] = v
	}

	return t.tfClient.SetEnv(envVars)
}

// initModule initializes terraform in the working directory
func (t *localTerraformExecutor) initModule(ctx context.Context) error {
	if err := t.tfClient.Init(ctx, tfexec.BackendConfig(backendConfigFilename)); err != nil {
		return err
	}

	return nil
}

// planModule runs terraform plan
func (t *localTerraformExecutor) planModule(ctx context.Context) error {
	if _, err := t.tfClient.Plan(ctx, tfexec.Refresh(true), tfexec.VarFile(varsFilename)); err != nil {
		return err
	}

	return nil
}

// applyModule runs terraform apply for the current module
func (t *localTerraformExecutor) applyModule(ctx context.Context) error {
	if err := t.tfClient.Apply(ctx, tfexec.Refresh(true), tfexec.VarFile(varsFilename)); err != nil {
		return err
	}

	return nil
}

// destroyModule runs terraform apply -destroy for the current module
func (t *localTerraformExecutor) destroyModule(ctx context.Context) error {
	if err := t.tfClient.Destroy(ctx, tfexec.Refresh(true), tfexec.VarFile(varsFilename)); err != nil {
		return err
	}

	return nil
}

type outputter interface {
	Output(context.Context, ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
}

var _ outputter = (*tfexec.Terraform)(nil)

func (t *localTerraformExecutor) outputs(ctx context.Context) (map[string]interface{}, error) {
	if t.outputter == nil {
		return nil, fmt.Errorf("missing outputter implementation")
	}
	m := map[string]interface{}{}

	out, err := t.outputter.Output(ctx)
	if err != nil {
		return m, err
	}

	for k, v := range out {
		var val interface{}
		err = json.Unmarshal(v.Value, &val)
		if err != nil {
			return m, err
		}
		m[k] = val
	}

	return m, err
}
