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
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
)

const updatePlanFilename = ".pulumi-update-plan.json"

const stateUploadTimeout = 60 * time.Second

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	plan := h.state.plan
	backend := plan.PulumiBackend
	workDir := h.state.srcWorkspace.Root()

	// Tag this handler's logger with semantic-convention attributes so every
	// emitted record (including from helpers further down the call tree) carries
	// them automatically.
	pulumiWorkspaceID := ""
	pulumiStack := ""
	if backend != nil {
		pulumiWorkspaceID = backend.WorkspaceID
		pulumiStack = backend.StackName
	}
	l = l.With(
		zap.String("service.name", "runner.sandbox.pulumi"),
		zap.String("nuon.tool", "pulumi"),
		zap.String("nuon.deploy.kind", "sandbox.pulumi"),
		zap.String("pulumi.workspace_id", pulumiWorkspaceID),
		zap.String("pulumi.stack", pulumiStack),
		zap.String("pulumi.operation", string(job.Operation)),
	)
	ctx = pkgctx.SetLogger(ctx, l)

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

	ws, err := pulumiworkspace.New(ctx, &pulumiworkspace.Options{
		WorkDir:   workDir,
		StackName: backend.StackName,
		Runtime:   backend.Runtime,
		Config:    backend.Config,
		EnvVars:   envVars,
		Logger:    l,
		StateBackend: &pulumiworkspace.StateBackend{
			APIEndpoint: h.cfg.RunnerAPIURL,
			WorkspaceID: backend.WorkspaceID,
			Token:       h.cfg.RunnerAPIToken,
			JobID:       h.state.jobID,
		},
	})
	if err != nil {
		h.writeErrorResult(ctx, "create pulumi workspace", err)
		return fmt.Errorf("unable to create pulumi workspace: %w", err)
	}
	h.state.workspace = ws

	if _, err := h.downloadState(ctx, l, ws, backend.WorkspaceID); err != nil {
		h.writeErrorResult(ctx, "download pulumi state", err)
		return fmt.Errorf("unable to download pulumi state: %w", err)
	}

	usePlans := backend.UpdatePlans

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		previewOpts := &pulumiworkspace.PreviewOpts{}
		if usePlans {
			previewOpts.PlanOutPath = filepath.Join(workDir, updatePlanFilename)
		}
		l.Info("executing pulumi preview")
		result, err := ws.Preview(ctx, previewOpts)
		if err != nil {
			l.Error("pulumi preview errored", zap.Error(err))
			h.writeErrorResult(ctx, "pulumi preview", err)
			return fmt.Errorf("unable to execute pulumi preview: %w", err)
		}

		var bundle []byte
		if previewOpts.PlanOutPath != "" {
			if bundle, err = h.bundleUpdatePlan(ctx, ws, previewOpts.PlanOutPath); err != nil {
				l.Warn("unable to bundle saved pulumi plan", zap.Error(err))
			}
		}

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
		// Persist state on every exit (success, error, or panic) so partially
		// created resources are never lost — the retry then reconciles instead
		// of recreating.
		defer func() {
			if err := h.updatePulumiState(ctx, ws); err != nil {
				l.Error("failed to persist pulumi state", zap.Error(err))
			}
		}()

		if isDeprovisionJob(job) {
			l.Info("executing pulumi destroy")
			if err := ws.Destroy(ctx); err != nil {
				l.Error("pulumi destroy errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi destroy", err)
				return fmt.Errorf("unable to execute pulumi destroy: %w", err)
			}
			l.Info("pulumi destroy completed")
		} else {
			upOpts := &pulumiworkspace.UpOpts{}
			if usePlans && h.state.plan.ApplyPlanContents != "" {
				planPath, err := h.materializeUpdatePlan(ctx, ws, h.state.plan.ApplyPlanContents)
				if err != nil {
					l.Warn("unable to materialize saved pulumi plan, falling back to fresh diff", zap.Error(err))
				} else {
					upOpts.PlanInPath = planPath
					l.Info("applying update plan saved by preview job", zap.String("plan_path", planPath))
				}
			}

			l.Info("executing pulumi up")
			result, err := ws.Up(ctx, upOpts)
			if err != nil {
				l.Error("pulumi up errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi up", err)
				return fmt.Errorf("unable to execute pulumi up: %w", err)
			}

			h.state.outputs = result.Outputs
			l.Info("pulumi up completed", zap.Any("outputs", result.Outputs))
		}

	default:
		return fmt.Errorf("unsupported operation type %s", job.Operation)
	}

	return nil
}

func isDeprovisionJob(job *models.AppRunnerJob) bool {
	if job == nil || job.Metadata == nil {
		return false
	}
	return job.Metadata["sandbox_run_type"] == "deprovision"
}

func (h *handler) writePlanResult(ctx context.Context, result *pulumiworkspace.PreviewResult, planFileBytes []byte) error {
	displayJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("unable to marshal preview result: %w", err)
	}
	displayB64, err := gzipBase64URL(displayJSON)
	if err != nil {
		return fmt.Errorf("unable to gzip preview result: %w", err)
	}

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

	planPath := filepath.Join(h.state.srcWorkspace.Root(), updatePlanFilename)
	if err := os.WriteFile(planPath, planJSON, 0o600); err != nil {
		return "", fmt.Errorf("unable to write plan file: %w", err)
	}
	return planPath, nil
}

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
	// Detach from job cancellation so a mid-deploy cancel still exports + persists
	// state; otherwise created resources are dropped and the retry recreates them.
	ctx = context.WithoutCancel(ctx)

	stateJSON, err := ws.ExportState(ctx)
	if err != nil {
		return fmt.Errorf("unable to export pulumi state: %w", err)
	}

	return uploadPulumiState(ctx, h.cfg.RunnerAPIURL, h.cfg.RunnerAPIToken, h.state.plan.PulumiBackend.WorkspaceID, h.state.jobID, stateJSON)
}

func uploadPulumiState(ctx context.Context, apiURL, token, workspaceID, jobID string, stateJSON []byte) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stateUploadTimeout)
	defer cancel()

	stateURL := fmt.Sprintf("%s/v1/runners/pulumi-state/%s?job_id=%s", apiURL, workspaceID, jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, stateURL, bytes.NewReader(stateJSON))
	if err != nil {
		return fmt.Errorf("unable to create state upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+token)

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
