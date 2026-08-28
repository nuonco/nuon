package cmd

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

func waitForInstallControls(ctx context.Context, store statestore.Store, deps *customerManagedOperationDeps, client *customermanaged.Client, logger *zap.Logger) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.Done():
			return
		case <-ticker.C:
			runID := client.RunID()
			for _, action := range []string{statestore.ControlActionCancel, statestore.ControlActionUserSkip, statestore.ControlActionRetry} {
				handledKey := statestore.InstallControlHandledKey(runID, action)
				if _, handled, _ := store.ReadFile(handledKey); handled {
					continue
				}
				requestKey := statestore.InstallControlKey(runID, action)
				raw, found, err := store.ReadFile(requestKey)
				if err != nil {
					logger.Error("read install control", zap.Error(err))
					continue
				}
				if !found && deps != nil {
					raw, found, err = deps.syncer.readControlObject(ctx, requestKey)
					if err != nil {
						logger.Error("read install control", zap.Error(err))
						continue
					}
				}
				if !found {
					continue
				}
				var request statestore.ControlRequest
				if err := json.Unmarshal(raw, &request); err != nil || request.RunID != runID || request.Action != action {
					logger.Error("invalid install control", zap.String("key", requestKey), zap.Error(err))
					continue
				}
				if err := client.ApplyControl(action); err != nil {
					logger.Error("apply install control", zap.String("action", action), zap.Error(err))
					continue
				}
				receipt, _ := json.Marshal(statestore.ControlHandled{ControlRequest: request, HandledAt: time.Now().UTC()})
				if err := store.WriteFile(handledKey, receipt); err != nil {
					logger.Error("write install control receipt", zap.Error(err))
				}
				logger.Info("install control applied", zap.String("run_id", runID), zap.String("action", action))
				break
			}
		}
	}
}

func waitForCandidateApproval(ctx context.Context, syncer *customerManagedS3Sync, digest string, client *customermanaged.Client, logger *zap.Logger) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.Done():
			return
		case <-ticker.C:
			required, phase := client.Approval()
			if !required {
				continue
			}
			approvalKey := operation.CandidateApprovalKey(digest)
			if phase == "sandbox" {
				approvalKey = operation.CandidateSandboxApprovalKey(digest)
			}
			_, found, err := syncer.readControlOperationObject(ctx, approvalKey)
			if err != nil || !found {
				continue
			}
			if err := client.ApproveApply(); err != nil {
				logger.Error("approve bundle candidate apply", zap.Error(err))
				continue
			}
			logger.Info("bundle candidate approved; continuing apply", zap.String("bundle_digest", digest), zap.String("phase", phase))
		}
	}
}
