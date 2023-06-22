package workspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace/output"
)

func (w *workspace) Init(ctx context.Context, log hclog.Logger) error {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return err
	}

	return w.init(ctx, client)
}

func (w *workspace) init(ctx context.Context, client Terraform) error {
	if err := client.Init(ctx,
		tfexec.BackendConfig(defaultBackendConfigFilename),
	); err != nil {
		return fmt.Errorf("unable to init terraform: %w", err)
	}

	return nil
}

func (w *workspace) Apply(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.apply(ctx, client, log)
}

func (w *workspace) apply(ctx context.Context, client Terraform, log hclog.Logger) ([]byte, error) {
	out, err := output.New(w.v, output.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("unable to get output: %w", err)
	}

	writer, err := out.Writer()
	if err != nil {
		return nil, fmt.Errorf("unable to get writer: %w", err)
	}

	if err := client.ApplyJSON(ctx,
		writer,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("error running apply: %w", err)
	}

	return out.Bytes()
}

func (w *workspace) Destroy(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.destroy(ctx, client, log)
}

func (w *workspace) destroy(ctx context.Context, client Terraform, log hclog.Logger) ([]byte, error) {
	out, err := output.New(w.v, output.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("unable to get output: %w", err)
	}

	writer, err := out.Writer()
	if err != nil {
		return nil, fmt.Errorf("unable to get writer: %w", err)
	}
	if err := client.DestroyJSON(ctx,
		writer,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("error running destroy: %w", err)
	}

	return out.Bytes()
}

func (w *workspace) Plan(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.plan(ctx, client, log)
}

func (w *workspace) plan(ctx context.Context, client Terraform, log hclog.Logger) ([]byte, error) {
	out, err := output.New(w.v, output.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("unable to get output: %w", err)
	}

	writer, err := out.Writer()
	if err != nil {
		return nil, fmt.Errorf("unable to get writer: %w", err)
	}
	if _, err := client.PlanJSON(ctx,
		writer,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("unable to plan: %w", err)
	}

	return out.Bytes()
}

func (w *workspace) Refresh(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.refresh(ctx, client, log)
}

func (w *workspace) refresh(ctx context.Context, client Terraform, log hclog.Logger) ([]byte, error) {
	out, err := output.New(w.v, output.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("unable to get output: %w", err)
	}

	writer, err := out.Writer()
	if err != nil {
		return nil, fmt.Errorf("unable to get writer: %w", err)
	}

	if err := client.RefreshJSON(ctx,
		writer,
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("unable to execute refresh: %w", err)
	}

	return out.Bytes()
}

func (w *workspace) Output(ctx context.Context, log hclog.Logger) (map[string]tfexec.OutputMeta, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.output(ctx, client)
}

func (w *workspace) output(ctx context.Context, client Terraform) (map[string]tfexec.OutputMeta, error) {
	out, err := client.Output(ctx)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (w *workspace) Show(ctx context.Context, log hclog.Logger) (*tfjson.State, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.show(ctx, client)
}

func (w *workspace) show(ctx context.Context, client Terraform) (*tfjson.State, error) {
	out, err := client.Show(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to execute show: %w", err)
	}

	return out, nil
}
