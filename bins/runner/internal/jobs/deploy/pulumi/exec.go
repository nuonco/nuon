package pulumi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.uber.org/zap"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/pkg/kube/config"
	pulumiworkspace "github.com/nuonco/nuon/pkg/pulumi/workspace"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return err
	}

	plan := h.state.plan.PulumiDeployPlan

	// Set up cloud auth env vars
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

	// Write kube config if cluster info is available
	if plan.ClusterInfo != nil {
		kubeConfigPath := filepath.Join(h.state.arch.BasePath(), config.DefaultKubeConfigFilename)
		if err := config.WriteConfig(ctx, plan.ClusterInfo, kubeConfigPath); err != nil {
			h.writeErrorResult(ctx, "write kube config", err)
			return fmt.Errorf("unable to write kube config: %w", err)
		}
		envVars["KUBECONFIG"] = kubeConfigPath
	}

	// Create Pulumi workspace
	ws, err := pulumiworkspace.New(ctx, &pulumiworkspace.Options{
		WorkDir:   h.state.arch.BasePath(),
		StackName: plan.StackName,
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

	// Download existing state from control plane and import into local backend
	if err := h.downloadState(ctx, l, ws, plan.WorkspaceID); err != nil {
		h.writeErrorResult(ctx, "download pulumi state", err)
		return fmt.Errorf("unable to download pulumi state: %w", err)
	}

	switch job.Operation {
	case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
		l.Info("executing pulumi preview")
		result, err := ws.Preview(ctx)
		if err != nil {
			l.Error("pulumi preview errored", zap.Error(err))
			h.writeErrorResult(ctx, "pulumi preview", err)
			return fmt.Errorf("unable to execute pulumi preview: %w", err)
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			h.writeErrorResult(ctx, "marshal preview result", err)
			return fmt.Errorf("unable to marshal preview result: %w", err)
		}
		resultB64 := base64.URLEncoding.EncodeToString(resultJSON)

		if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			Success:            true,
			ContentsCompressed: resultB64,
		}); err != nil {
			h.errRecorder.Record("write job execution result", err)
		}

	case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
		l.Info("executing pulumi destroy preview")
		result, err := ws.Preview(ctx)
		if err != nil {
			l.Error("pulumi destroy preview errored", zap.Error(err))
			h.writeErrorResult(ctx, "pulumi destroy preview", err)
			return fmt.Errorf("unable to execute pulumi destroy preview: %w", err)
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			h.writeErrorResult(ctx, "marshal preview result", err)
			return fmt.Errorf("unable to marshal preview result: %w", err)
		}
		resultB64 := base64.URLEncoding.EncodeToString(resultJSON)

		if _, err := h.apiClient.CreateJobExecutionResult(ctx, h.state.jobID, h.state.jobExecutionID, &models.ServiceCreateRunnerJobExecutionResultRequest{
			Success:            true,
			ContentsCompressed: resultB64,
		}); err != nil {
			h.errRecorder.Record("write job execution result", err)
		}

	case models.AppRunnerJobOperationTypeApplyDashPlan:
		if plan.Destroy {
			l.Info("executing pulumi destroy")
			if err := ws.Destroy(ctx); err != nil {
				l.Error("pulumi destroy errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi destroy", err)
				// Still try to upload state even on failure
				if stateErr := h.updatePulumiState(ctx, ws); stateErr != nil {
					l.Error("failed to update state after error", zap.Error(stateErr))
				}
				return fmt.Errorf("unable to execute pulumi destroy: %w", err)
			}

			// Update state in control plane after destroy
			if err := h.updatePulumiState(ctx, ws); err != nil {
				h.writeErrorResult(ctx, "update pulumi state", err)
			}

			l.Info("pulumi destroy completed")
		} else {
			l.Info("executing pulumi up")
			result, err := ws.Up(ctx)
			if err != nil {
				l.Error("pulumi up errored", zap.Error(err))
				h.writeErrorResult(ctx, "pulumi up", err)
				// Still try to upload state even on failure
				if stateErr := h.updatePulumiState(ctx, ws); stateErr != nil {
					l.Error("failed to update state after error", zap.Error(stateErr))
				}
				return fmt.Errorf("unable to execute pulumi up: %w", err)
			}

			// Update state in control plane
			if err := h.updatePulumiState(ctx, ws); err != nil {
				h.writeErrorResult(ctx, "update pulumi state", err)
			}

			l.Info("pulumi up completed", zap.Any("outputs", result.Outputs))
		}

	default:
		return fmt.Errorf("unsupported operation type %s", job.Operation)
	}

	return nil
}

func (h *handler) downloadState(ctx context.Context, l *zap.Logger, ws *pulumiworkspace.Workspace, workspaceID string) error {
	l.Info("downloading pulumi state from control plane", zap.String("workspace_id", workspaceID))

	// Fetch state directly via HTTP from the runner API
	stateURL := fmt.Sprintf("%s/v1/terraform-workspaces/%s/state-json", h.cfg.RunnerAPIURL, workspaceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stateURL, nil)
	if err != nil {
		return fmt.Errorf("unable to create state request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.cfg.RunnerAPIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Info("unable to fetch state, starting fresh", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		l.Info("no existing state found, starting fresh")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		l.Info("non-OK response fetching state, starting fresh", zap.Int("status", resp.StatusCode))
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read state response: %w", err)
	}

	// Skip empty or null state
	if len(body) == 0 || string(body) == "null" || string(body) == "[]" {
		l.Info("state is empty, starting fresh")
		return nil
	}

	// The response is an array of workspace states, extract the latest state contents
	var states []json.RawMessage
	if err := json.Unmarshal(body, &states); err != nil {
		// Try parsing as a single state object
		l.Info("importing existing state into local backend", zap.Int("state_bytes", len(body)))
		if err := ws.ImportState(ctx, body); err != nil {
			return fmt.Errorf("unable to import state: %w", err)
		}
		return nil
	}

	if len(states) == 0 {
		l.Info("no states found, starting fresh")
		return nil
	}

	// Parse the first (latest) state to get its contents field
	var stateRecord struct {
		Contents json.RawMessage `json:"contents"`
	}
	if err := json.Unmarshal(states[0], &stateRecord); err != nil {
		return fmt.Errorf("unable to parse state record: %w", err)
	}

	if len(stateRecord.Contents) == 0 || string(stateRecord.Contents) == "null" {
		l.Info("state contents empty, starting fresh")
		return nil
	}

	l.Info("importing existing state into local backend", zap.Int("state_bytes", len(stateRecord.Contents)))
	if err := ws.ImportState(ctx, stateRecord.Contents); err != nil {
		return fmt.Errorf("unable to import state: %w", err)
	}

	return nil
}

func (h *handler) updatePulumiState(ctx context.Context, ws *pulumiworkspace.Workspace) error {
	stateJSON, err := ws.ExportState(ctx)
	if err != nil {
		return fmt.Errorf("unable to export pulumi state: %w", err)
	}

	if _, err := h.apiClient.UpdateTerraformStateJSON(
		ctx,
		h.state.plan.PulumiDeployPlan.WorkspaceID,
		&h.state.jobID,
		stateJSON,
	); err != nil {
		return fmt.Errorf("unable to update state: %w", err)
	}

	return nil
}
