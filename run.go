package terraform

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/go-playground/validator/v10"
)

type RunType string

const (
	RunTypePlanAndApply RunType = "plan_and_apply"
	RunTypePlanOnly     RunType = "plan_only"
	RunTypeDestroy      RunType = "destroy"
)

type RunRequest struct {
	ID      string  `validate:"required"`
	RunType RunType `validate:"required"`

	Module Module `validate:"required"`

	Stdout io.Writer `validate:"required"`
	Stderr io.Writer `validate:"required"`

	BackendConfig BackendConfig          `validate:"required"`
	EnvVars       map[string]string      `validate:"required"`
	TfVars        map[string]interface{} `validate:"required"`
}

func (r RunRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type RunResponse struct {
	RunType RunType
	Output  map[string]interface{}
}

type run struct {
	workspace         terraformWorkspace
	backendConfigurer backendConfigurer
	varsConfigurer    varsConfigurer
	terraformExecutor terraformExecutor
}

func (r *run) run(ctx context.Context, req RunRequest) (RunResponse, error) {
	var (
		resp RunResponse
		err  error
	)

	logger := log.New(req.Stdout, "terraform: ", log.Lshortfile)
	if err = r.workspace.init(ctx, logger); err != nil {
		return resp, fmt.Errorf("unable to init workspace: %w", err)
	}
	defer func() {
		err = r.workspace.cleanup(context.Background())
	}()

	if err = r.backendConfigurer.createBackendConfig(req.BackendConfig, r.workspace); err != nil {
		return resp, fmt.Errorf("unable to create backend config: %w", err)
	}

	if err = r.varsConfigurer.createVarsConfigFile(req.TfVars, r.workspace); err != nil {
		return resp, fmt.Errorf("unable to create vars config file: %w", err)
	}

	// initialize and execute terraform
	if err = r.terraformExecutor.initClient(r.workspace.getTmpDir(), r.workspace.getTfExecPath()); err != nil {
		return resp, fmt.Errorf("unable to iniitalize terraform: %w", err)
	}
	r.terraformExecutor.setLogger(logger)
	r.terraformExecutor.setStdout(req.Stdout)
	r.terraformExecutor.setStderr(req.Stderr)
	if err = r.terraformExecutor.initModule(ctx); err != nil {
		return resp, fmt.Errorf("unable to initialize module: %w", err)
	}

	if err = r.handleRunType(ctx, req.RunType); err != nil {
		return resp, err
	}

	outputs, err := r.terraformExecutor.outputs(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get outputs: %w", err)
	}
	resp.Output = outputs
	return resp, nil
}

func (r *run) handleRunType(ctx context.Context, typ RunType) error {
	if typ == RunTypeDestroy {
		if err := r.terraformExecutor.destroyModule(ctx); err != nil {
			return fmt.Errorf("unable to run destroy: %w", err)
		}
		return nil
	}

	// NOTE: if it is not a run type of destroy, we always run plan
	if err := r.terraformExecutor.planModule(ctx); err != nil {
		return fmt.Errorf("unable to run plan during %s: %w", typ, err)
	}

	if typ == RunTypePlanOnly {
		return nil
	}

	if err := r.terraformExecutor.applyModule(ctx); err != nil {
		return fmt.Errorf("unable to run apply: %w", err)
	}

	return nil
}

func Run(ctx context.Context, req RunRequest) (RunResponse, error) {
	var resp RunResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	run := &run{
		workspace: &workspace{
			ID:            req.ID,
			module:        req.Module,
			installer:     &tfInstaller{},
			moduleFetcher: &s3ModuleFetcher{},
		},
		backendConfigurer: &s3BackendConfigurer{},
		varsConfigurer:    &tfVarsConfigurer{},
		terraformExecutor: &localTerraformExecutor{},
	}

	return run.run(ctx, req)
}
