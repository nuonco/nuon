package pulumi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/kube/config"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
)

const updatePlanFilename = ".pulumi-update-plan.json"

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	plan := h.state.plan.PulumiDeployPlan

	envVars := make(map[string]string)
	for k, v := range plan.EnvVars {
		envVars[k] = v
	}

	if plan.AWSAuth != nil {
		awsEnvVars, err := awscredentials.FetchEnv(ctx, plan.AWSAuth)
		if err != nil {
			h.writeErrorResult(ctx, "fetch aws credentials", err)
			return fmt.Errorf("unable to fetch AWS credentials: %w", err)
		}
		for k, v := range awsEnvVars {
			envVars[k] = v
		}
	}

	if plan.AzureAuth != nil {
		azureEnvVars, err := azurecredentials.FetchEnv(ctx, plan.AzureAuth)
		if err != nil {
			h.writeErrorResult(ctx, "fetch azure credentials", err)
			return fmt.Errorf("unable to fetch Azure credentials: %w", err)
		}
		for k, v := range azureEnvVars {
			envVars[k] = v
		}
	}

	if plan.GCPAuth != nil {
		gcpEnvVars, err := gcpcredentials.FetchEnv(ctx, plan.GCPAuth)
		if err != nil {
			h.writeErrorResult(ctx, "fetch gcp credentials", err)
			return fmt.Errorf("unable to fetch GCP credentials: %w", err)
		}
		for k, v := range gcpEnvVars {
			envVars[k] = v
		}
	}

	if plan.ClusterInfo != nil {
		kubeConfigPath := filepath.Join(h.state.arch.BasePath(), config.DefaultKubeConfigFilename)
		if err := config.WriteConfig(ctx, plan.ClusterInfo, kubeConfigPath); err != nil {
			h.writeErrorResult(ctx, "write kube config", err)
			return fmt.Errorf("unable to write kube config: %w", err)
		}
		envVars["KUBECONFIG"] = kubeConfigPath
	}

	ws, err := pulumiworkspace.New(ctx, &pulumiworkspace.Options{
		WorkDir:   h.state.arch.BasePath(),
		StackName: plan.StackName,
		Runtime:   plan.Runtime,
		Config:    plan.Config,
		EnvVars:   envVars,
		StateBackend: &pulumiworkspace.StateBackend{
			APIEndpoint: h.cfg.RunnerAPIURL,
			WorkspaceID: plan.WorkspaceID,
			Token:       h.cfg.RunnerAPIToken,
			JobID:       h.state.jobID,
		},
	})
	if err != nil {
		h.writeErrorResult(ctx, "create pulumi workspace", err)
		return fmt.Errorf("unable to create pulumi workspace: %w", err)
	}
	h.state.workspace = ws

	if _, err := h.downloadState(ctx, l, ws, plan.WorkspaceID); err != nil {
		h.writeErrorResult(ctx, "download pulumi state", err)
		return fmt.Errorf("unable to download pulumi state: %w", err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		planOutPath := filepath.Join(h.state.arch.BasePath(), updatePlanFilename)
		l.Info("executing pulumi preview", zap.String("plan_out", planOutPath))
		result, err := ws.Preview(ctx, &pulumiworkspace.PreviewOpts{PlanOutPath: planOutPath})
		if err != nil {
			l.Error("pulumi preview errored", zap.Error(err))
			h.writeErrorResult(ctx, "pulumi preview", err)
			return fmt.Errorf("unable to execute pulumi preview: %w", err)
		}

		bundle, err := h.bundleUpdatePlan(ctx, ws, planOutPath)
		if err != nil {
			l.Warn("unable to bundle saved pulumi plan", zap.Error(err))
		}
		l.Info("saved update plan from preview, ready for apply job", zap.Int("bundle_bytes", len(bundle)))

		if err := h.writePlanResult(ctx, result, bundle); err != nil {
			h.errRecorder.Record("write job execution result", err)
		}

	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Info("executing pulumi destroy preview")
		result, err := ws.DestroyPreview(ctx)
		if err != nil {
			l.Error("pulumi destroy preview errored", zap.Error(err))
			h.writeErrorResult(ctx, "pulumi destroy preview", err)
			return fmt.Errorf("unable to execute pulumi destroy preview: %w", err)
		}

		if err := h.writePlanResult(ctx, result, nil); err != nil {
			h.errRecorder.Record("write job execution result", err)
		}

	case models.AppRunnerJobOperationTypeApplyDashPlan:
		if plan.Destroy {
			l.Info("executing pulumi destroy")
			if err := ws.Destroy(ctx); err != nil {
				l.Error("pulumi destroy errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi destroy", err)
				if stateErr := h.updatePulumiState(ctx, ws); stateErr != nil {
					l.Error("failed to update state after error", zap.Error(stateErr))
				}
				return fmt.Errorf("unable to execute pulumi destroy: %w", err)
			}

			if err := h.updatePulumiState(ctx, ws); err != nil {
				h.writeErrorResult(ctx, "update pulumi state", err)
			}

			l.Info("pulumi destroy completed")
		} else {
			upOpts := &pulumiworkspace.UpOpts{}
			if h.state.plan.ApplyPlanContents != "" {
				planPath, err := h.materializeUpdatePlan(ctx, ws, h.state.plan.ApplyPlanContents)
				if err != nil {
					l.Warn("unable to materialize saved pulumi plan, falling back to fresh diff", zap.Error(err))
				} else {
					upOpts.PlanInPath = planPath
					l.Info("applying update plan saved by preview job", zap.String("plan_path", planPath))
				}
			} else {
				l.Info("no update plan from preview job, computing fresh diff at apply time")
			}

			l.Info("executing pulumi up")
			result, err := ws.Up(ctx, upOpts)
			if err != nil {
				l.Error("pulumi up errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi up", err)
				if stateErr := h.updatePulumiState(ctx, ws); stateErr != nil {
					l.Error("failed to update state after error", zap.Error(stateErr))
				}
				return fmt.Errorf("unable to execute pulumi up: %w", err)
			}

			if err := h.updatePulumiState(ctx, ws); err != nil {
				h.writeErrorResult(ctx, "update pulumi state", err)
			}

			h.state.outputs = result.Outputs
			l.Info("pulumi up completed", zap.Any("outputs", result.Outputs))
		}

	default:
		return fmt.Errorf("unsupported operation type %s", job.Operation)
	}

	return nil
}

// writePlanResult uploads two payloads on the job execution result:
//   - ContentsCompressed: the Pulumi update plan file (gzip+b64), used by the
//     subsequent apply job to skip its own preview and enforce drift safety.
//   - ContentsDisplayCompressed: the structured PreviewResult JSON (gzip+b64),
//     used by the dashboard to render the per-resource diff.
func (h *handler) writePlanResult(ctx context.Context, result *pulumiworkspace.PreviewResult, planFileBytes []byte) error {
	displayJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("unable to marshal preview result: %w", err)
	}
	displayB64, err := gzipBase64URL(displayJSON)
	if err != nil {
		return fmt.Errorf("unable to gzip preview result: %w", err)
	}

	// Orchestrator requires non-empty contents for non-NOOP jobs; teardowns have no real plan.
	contentsB64 := displayB64
	if len(planFileBytes) > 0 {
		contentsB64, err = gzipBase64URL(planFileBytes)
		if err != nil {
			return fmt.Errorf("unable to gzip update plan: %w", err)
		}
	}

	if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:                   true,
		ContentsCompressed:        contentsB64,
		ContentsDisplayCompressed: displayB64,
	}); err != nil {
		return fmt.Errorf("unable to create job execution result: %w", err)
	}

	return nil
}

func gzipBase64URL(raw []byte) (string, error) {
	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(raw); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(gzBuf.Bytes()), nil
}

// updatePlanBundle pairs the plan job's stack encryption salt with the saved
// plan, so the apply job can decrypt secret values in the plan even on a
// fresh stack with no prior state.
type updatePlanBundle struct {
	Version int    `json:"v"`
	Salt    string `json:"salt,omitempty"`
	PlanB64 string `json:"plan_b64"`
}

func (h *handler) bundleUpdatePlan(ctx context.Context, ws *pulumiworkspace.Workspace, planPath string) ([]byte, error) {
	planJSON, err := os.ReadFile(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read plan file: %w", err)
	}
	if len(planJSON) == 0 {
		return nil, nil
	}

	salt, err := ws.EncryptionSalt(ctx)
	if err != nil {
		return nil, fmt.Errorf("read stack encryption salt: %w", err)
	}

	return json.Marshal(updatePlanBundle{
		Version: 1,
		Salt:    salt,
		PlanB64: base64.StdEncoding.EncodeToString(planJSON),
	})
}

// materializeUpdatePlan reverses bundleUpdatePlan + the gzip+b64 round-trip
// the API server performs: decode, decompress, restore the plan job's salt
// onto this stack, and write the plan JSON for Pulumi to consume via --plan.
func (h *handler) materializeUpdatePlan(ctx context.Context, ws *pulumiworkspace.Workspace, b64Contents string) (string, error) {
	gzBytes, err := base64.StdEncoding.DecodeString(b64Contents)
	if err != nil {
		return "", fmt.Errorf("unable to base64-decode plan bundle: %w", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(gzBytes))
	if err != nil {
		return "", fmt.Errorf("unable to open gzip reader for plan bundle: %w", err)
	}
	defer gzReader.Close()

	bundleJSON, err := io.ReadAll(gzReader)
	if err != nil {
		return "", fmt.Errorf("unable to read decompressed plan bundle: %w", err)
	}

	var bundle updatePlanBundle
	if err := json.Unmarshal(bundleJSON, &bundle); err != nil {
		return "", fmt.Errorf("unable to parse plan bundle: %w", err)
	}

	if bundle.PlanB64 == "" {
		return "", fmt.Errorf("plan payload missing bundle wrapper (older runner format?)")
	}

	if err := ws.SetEncryptionSalt(ctx, bundle.Salt); err != nil {
		return "", fmt.Errorf("unable to restore encryption salt: %w", err)
	}

	planJSON, err := base64.StdEncoding.DecodeString(bundle.PlanB64)
	if err != nil {
		return "", fmt.Errorf("unable to base64-decode plan: %w", err)
	}

	planPath := filepath.Join(h.state.arch.BasePath(), updatePlanFilename)
	if err := os.WriteFile(planPath, planJSON, 0o600); err != nil {
		return "", fmt.Errorf("unable to write plan file: %w", err)
	}
	return planPath, nil
}

// downloadState fetches the current pulumi state from the control plane and
// imports it. Returns true when state was imported. Update plans can't cross
// fresh-stack boundaries (each fresh stack gets its own encryption salt), so
// callers gate --save-plan / --plan on this signal.
func (h *handler) downloadState(ctx context.Context, l *zap.Logger, ws *pulumiworkspace.Workspace, workspaceID string) (bool, error) {
	l.Info("downloading pulumi state from control plane", zap.String("workspace_id", workspaceID))

	stateURL := fmt.Sprintf("%s/v1/runners/pulumi-state/%s",
		h.cfg.RunnerAPIURL, workspaceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stateURL, nil)
	if err != nil {
		return false, fmt.Errorf("unable to create state request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.cfg.RunnerAPIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Info("unable to fetch prior pulumi state — first-time deploy", zap.Error(err))
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		l.Info("no prior pulumi state in control plane — first-time deploy")
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		l.Info("non-OK response fetching prior pulumi state — first-time deploy", zap.Int("status", resp.StatusCode))
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("unable to read state response: %w", err)
	}

	if len(body) == 0 {
		l.Info("prior pulumi state is empty — first-time deploy")
		return false, nil
	}

	l.Info("importing prior pulumi state into local backend", zap.Int("state_bytes", len(body)))
	if err := ws.ImportState(ctx, body); err != nil {
		return false, fmt.Errorf("unable to import state: %w", err)
	}

	return true, nil
}

func (h *handler) updatePulumiState(ctx context.Context, ws *pulumiworkspace.Workspace) error {
	stateJSON, err := ws.ExportState(ctx)
	if err != nil {
		return fmt.Errorf("unable to export pulumi state: %w", err)
	}

	workspaceID := h.state.plan.PulumiDeployPlan.WorkspaceID
	stateURL := fmt.Sprintf("%s/v1/runners/pulumi-state/%s?job_id=%s",
		h.cfg.RunnerAPIURL, workspaceID, h.state.jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, stateURL, bytes.NewReader(stateJSON))
	if err != nil {
		return fmt.Errorf("unable to create state upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+h.cfg.RunnerAPIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to upload state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("state upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
