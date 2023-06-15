package workspace

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

// TODO(jm): rename this
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

	return w.apply(ctx, client)
}

func (w *workspace) apply(ctx context.Context, client Terraform) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := client.ApplyJSON(ctx,
		buf,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("error running apply: %w", err)
	}

	return buf.Bytes(), nil
}

func (w *workspace) Destroy(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.destroy(ctx, client)
}

func (w *workspace) destroy(ctx context.Context, client Terraform) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := client.DestroyJSON(ctx,
		buf,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("error running destroy: %w", err)
	}

	return buf.Bytes(), nil
}

func (w *workspace) Plan(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.plan(ctx, client)
}

func (w *workspace) plan(ctx context.Context, client Terraform) ([]byte, error) {
	if _, err := client.Plan(ctx,
		tfexec.Refresh(true),
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("unable to plan: %w", err)
	}

	return nil, nil
}

func (w *workspace) Refresh(ctx context.Context, log hclog.Logger) ([]byte, error) {
	client, err := w.getClient(ctx, log)
	if err != nil {
		return nil, err
	}

	return w.refresh(ctx, client)
}

func (w *workspace) refresh(ctx context.Context, client Terraform) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := client.RefreshJSON(ctx,
		buf,
		tfexec.VarFile(defaultVariablesFilename),
	); err != nil {
		return nil, fmt.Errorf("unable to execute refresh: %w", err)
	}

	return buf.Bytes(), nil
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
