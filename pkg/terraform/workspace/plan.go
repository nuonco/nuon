package workspace

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/terraform/workspace/output"
	"go.uber.org/zap"
)

const (
	defaultPlanFilename    = "tfplan"
	compressedPlanFilename = "tfplan.gz"
)

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

	opts := []tfexec.PlanOption{
		tfexec.Refresh(true),
		tfexec.Out("tfplan"), // NOTE: this should probably be configured w/ a WithPlanOut
	}
	for _, fp := range w.varsPaths {
		opts = append(opts, tfexec.VarFile(fp))
	}

	var diffExists bool
	if diffExists, err = client.PlanJSON(ctx,
		writer,
		opts...,
	); err != nil {
		fmt.Printf("%e", err)
		return nil, fmt.Errorf("unable to plan: %w", err)
	}

	log.Debug("plan diff", zap.Bool("diff.exists", diffExists))

	return out.Bytes()
}

// WritePlan writes the Terraform tfplan to a file called tfplan in the workspace. the bytes are provided externally. e.g. in the runner in exec.
func (w *workspace) WriteTFPlan(ctx context.Context, log hclog.Logger) ([]byte, error) {
	// NOTE: the plan is expected to be an opaque format tfplan file (not human legible)

	// Create the plan.json file in the workspace directory this method writes the raw bytes.

	// write the tfplan to a file in the workspace directory
	planFilePath := filepath.Join(w.root, defaultPlanFilename)
	log.Debug("writing plan", zap.String("path", planFilePath), zap.Int("plan.bytes.count", len(w.PlanBytes)))
	fd, err := os.Create(planFilePath)
	if err != nil {
		defer fd.Close()
		return []byte{}, fmt.Errorf("unable to create %s file: %w", defaultPlanFilename, err)
	}
	n, err := fd.Write(w.PlanBytes)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to write %s file: %w", defaultPlanFilename, err)
	}
	log.Debug("wrote plan", zap.String("path", planFilePath), zap.Int("plan.bytes.count", len(w.PlanBytes)), zap.Int("bytes-written", n))
	fd.Sync()

	// compress the tfplan file and write it as tfplan.gz
	var zipBytes bytes.Buffer
	gzipWriter := gzip.NewWriter(&zipBytes)
	gzipWriter.Write(w.PlanBytes)
	gzipWriter.Close()

	compressedPlanFilePath := filepath.Join(w.root, compressedPlanFilename)
	log.Debug("writing plan", zap.String("path", compressedPlanFilePath), zap.Int("plan.bytes.count", len(w.PlanBytes)))
	cfd, err := os.Create(compressedPlanFilePath)
	if err != nil {
		defer fd.Close()
		return []byte{}, fmt.Errorf("unable to create %s file: %w", defaultPlanFilename, err)
	}
	n, err = cfd.Write(zipBytes.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("unable to write %s file: %w", defaultPlanFilename, err)
	}
	log.Debug("wrote compressed plan", zap.String("path", compressedPlanFilePath), zap.Int("plan.bytes.count", len(zipBytes.Bytes())), zap.Int("bytes-written", n))
	fd.Sync()

	return w.PlanBytes, nil
}

// Compresses an existing tfplan already at the root
func (w *workspace) CompressTFPlan(ctx context.Context, log hclog.Logger) ([]byte, error) {

	planFilePath := filepath.Join(w.root, defaultPlanFilename)
	byts, err := os.ReadFile(planFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read tfplan: %w", err)
	}

	var zipBytes bytes.Buffer
	gzipWriter := gzip.NewWriter(&zipBytes)
	gzipWriter.Write(byts)
	gzipWriter.Close()

	compressedPlanFilePath := filepath.Join(w.root, compressedPlanFilename)
	log.Debug("writing plan", zap.String("path", compressedPlanFilePath), zap.Int("plan.bytes.count", len(byts)))
	cfd, err := os.Create(compressedPlanFilePath)
	if err != nil {
		defer cfd.Close()
		return []byte{}, fmt.Errorf("unable to create %s file: %w", defaultPlanFilename, err)
	}
	n, err := cfd.Write(zipBytes.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("unable to write %s file: %w", defaultPlanFilename, err)
	}
	log.Debug("wrote compressed plan", zap.String("path", compressedPlanFilePath), zap.Int("plan.bytes.count", len(zipBytes.Bytes())), zap.Int("bytes-written", n))
	cfd.Sync()

	return w.PlanBytes, nil
}

func (w *workspace) GetTfplan(ctx context.Context, log hclog.Logger) ([]byte, error) {
	bytes, err := os.ReadFile(filepath.Join(w.root, defaultPlanFilename))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read tfplan")
	}
	return bytes, nil
}

func (w *workspace) GetTfplanCompressed(ctx context.Context, log hclog.Logger) ([]byte, error) {
	bytes, err := os.ReadFile(filepath.Join(w.root, compressedPlanFilename))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read tfplan.gz")
	}
	return bytes, nil
}

func (w *workspace) GetTfplanJsonCompressed(ctx context.Context, log hclog.Logger) ([]byte, error) {
	bytes, err := os.ReadFile(filepath.Join(w.root, "plan.json.gz")) // TODO(fd): make this a const
	if err != nil {
		return nil, errors.Wrap(err, "unable to read plan.json.gz")
	}
	return bytes, nil
}
