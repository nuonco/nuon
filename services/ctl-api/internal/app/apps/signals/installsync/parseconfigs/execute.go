package parseconfigs

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	installsConfig, err := branchactivities.AwaitGetAppInstallsConfigByAppID(ctx, s.AppID)
	if err != nil {
		return fmt.Errorf("unable to get app installs config: %w", err)
	}
	if !installsConfig.Found {
		s.updateStepMetadata(ctx, map[string]any{"description": "no installs config found"})
		return nil
	}

	vcsConnectionID := ""
	if installsConfig.VCSConnectionID != nil {
		vcsConnectionID = *installsConfig.VCSConnectionID
	}

	cloneResult, err := branchactivities.AwaitCloneInstallsRepo(ctx, &branchactivities.CloneInstallsRepoInput{
		AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		VCSType:                installsConfig.VCSType,
		VCSConnectionID:        vcsConnectionID,
		Repo:                   installsConfig.Repo,
		Branch:                 installsConfig.Branch,
		CommitSHA:              s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone installs repo: %w", err)
	}

	installsDir := s.InstallsDirectory
	if installsDir == "" {
		installsDir = "."
	}

	parsedConfigs, err := branchactivities.AwaitParseInstallConfigs(ctx, branchactivities.ParseInstallConfigsRequest{
		SourceDir: cloneResult.SourceDir,
		Req: &branchactivities.ParseInstallConfigsInput{
			SourceDir:         cloneResult.SourceDir,
			InstallsDirectory: installsDir,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to parse install configs: %w", err)
	}

	if len(parsedConfigs.Installs) == 0 {
		s.updateStepMetadata(ctx, map[string]any{
			"total_configs": 0,
			"description":   "no install configs found",
		})
		logger.Info("no install configs found in repo")
		return nil
	}

	installs, err := branchactivities.AwaitGetInstallsForAppByAppID(ctx, s.AppID)
	if err != nil {
		return fmt.Errorf("unable to get installs for app: %w", err)
	}

	existingByName := make(map[string]struct{}, len(installs))
	for _, inst := range installs {
		existingByName[inst.Name] = struct{}{}
	}

	var missingInstalls []app.ProposedInstall
	for _, pc := range parsedConfigs.Installs {
		if pc.Config == nil {
			continue
		}
		if _, ok := existingByName[pc.Config.Name]; !ok {
			configJSON, _ := json.Marshal(pc.Config)
			missingInstalls = append(missingInstalls, app.ProposedInstall{
				Name:     pc.Config.Name,
				FilePath: pc.FilePath,
				Config:   configJSON,
			})
		}
	}

	meta := map[string]any{
		"total_configs":    len(parsedConfigs.Installs),
		"existing":         len(installs),
		"missing_installs": len(missingInstalls),
	}

	if len(missingInstalls) > 0 {
		planJSON, _ := json.Marshal(missingInstalls)

		proposedNames := make([]string, len(missingInstalls))
		for i, mi := range missingInstalls {
			proposedNames[i] = mi.Name
		}
		meta["description"] = fmt.Sprintf("found %d configs, %d new installs need approval", len(parsedConfigs.Installs), len(missingInstalls))
		meta["proposed_installs"] = missingInstalls
		meta["proposed_install_names"] = proposedNames
		s.updateStepMetadata(ctx, meta)

		_, err := branchactivities.AwaitCreateStepApproval(ctx, &branchactivities.CreateStepApprovalInput{
			OwnerID:   s.AppInstallConfigSyncID,
			OwnerType: "app_install_config_syncs",
			StepID:    s.StepID,
			Type:      app.InstallCreationApprovalType,
			Plan:      string(planJSON),
		})
		if err != nil {
			return fmt.Errorf("unable to create step approval: %w", err)
		}

		logger.Info(meta["description"].(string))
		return nil
	}

	meta["description"] = fmt.Sprintf("found %d configs, all installs exist", len(parsedConfigs.Installs))
	s.updateStepMetadata(ctx, meta)
	logger.Info(meta["description"].(string))
	return nil
}

func (s *Signal) updateStepMetadata(ctx workflow.Context, metadata map[string]any) {
	if s.StepID == "" {
		return
	}
	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:   app.StatusInProgress,
			Metadata: metadata,
		},
	})
}
