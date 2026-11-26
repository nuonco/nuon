package helpers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestCreateWorkflowGroupInstallApprovalOption(t *testing.T) {
	db := appliedHistoryPostgres(t)
	require.NoError(t, db.AutoMigrate(&app.InstallConfig{}, &app.AppBranchInstallGroup{}))
	helpers := installhelpers.New(installhelpers.Params{DB: db, L: zap.NewNop()})

	orgID := historyID("org", 0)
	accountID := historyID("account", 0)
	ctx := cctx.SetAccountIDContext(cctx.SetOrgIDContext(context.Background(), orgID), accountID)

	rollout := func(groupID string) map[string]string {
		metadata := map[string]string{app.WorkflowMetadataKeyStackChanged: "true"}
		if groupID != "" {
			metadata["install_group_id"] = groupID
		}
		return metadata
	}
	inputsOnly := func(groupID string) map[string]string {
		return map[string]string{
			app.WorkflowMetadataKeyInputsOnly: "true",
			"install_group_id":                groupID,
		}
	}

	tests := map[string]struct {
		groupOption   app.InstallApprovalOption
		installOption app.InstallApprovalOption
		workflowType  app.WorkflowType
		metadata      func(groupID string) map[string]string
		unknownGroup  bool
		noGroupID     bool
		wantOption    app.InstallApprovalOption
		wantType      string
	}{
		"rollout group approve-all overrides install prompt": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionPrompt,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			wantOption:    app.InstallApprovalOptionApproveAll,
			wantType:      "install-group",
		},
		"rollout group approve-all overrides install approve-all": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionApproveAll,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			wantOption:    app.InstallApprovalOptionApproveAll,
			wantType:      "install-group",
		},
		"rollout group prompt keeps install approve-all": {
			groupOption:   app.InstallApprovalOptionPrompt,
			installOption: app.InstallApprovalOptionApproveAll,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			wantOption:    app.InstallApprovalOptionApproveAll,
			wantType:      "install-config",
		},
		"rollout group prompt keeps install prompt": {
			groupOption:   app.InstallApprovalOptionPrompt,
			installOption: app.InstallApprovalOptionPrompt,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			wantOption:    app.InstallApprovalOptionPrompt,
		},
		"rollout without a group id uses the install config": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionPrompt,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			noGroupID:     true,
			wantOption:    app.InstallApprovalOptionPrompt,
		},
		"rollout with an unknown group id uses the install config": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionApproveAll,
			workflowType:  app.WorkflowTypeAppBranchConfigUpdate,
			metadata:      rollout,
			unknownGroup:  true,
			wantOption:    app.InstallApprovalOptionApproveAll,
			wantType:      "install-config",
		},
		"non-rollout workflow ignores a group approve-all": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionPrompt,
			workflowType:  app.WorkflowTypeInputUpdate,
			metadata:      inputsOnly,
			wantOption:    app.InstallApprovalOptionPrompt,
		},
		"non-rollout workflow keeps install approve-all": {
			groupOption:   app.InstallApprovalOptionApproveAll,
			installOption: app.InstallApprovalOptionApproveAll,
			workflowType:  app.WorkflowTypeInputUpdate,
			metadata:      inputsOnly,
			wantOption:    app.InstallApprovalOptionApproveAll,
			wantType:      "install-config",
		},
	}

	n := 0
	for name, tt := range tests {
		n++
		t.Run(name, func(t *testing.T) {
			install := app.Install{
				ID:          historyID("install", n),
				Name:        fmt.Sprintf("install-%d", n),
				OrgID:       orgID,
				CreatedByID: accountID,
				AppConfigID: historyID("config", n),
			}
			persistHistoryRow(t, db, &install)
			persistHistoryRow(t, db, &app.InstallConfig{
				ID:             historyID("instcfg", n),
				OrgID:          orgID,
				CreatedByID:    accountID,
				InstallID:      install.ID,
				ApprovalOption: tt.installOption,
			})
			group := app.AppBranchInstallGroup{
				ID:                    historyID("group", n),
				OrgID:                 orgID,
				CreatedByID:           accountID,
				AppBranchConfigID:     historyID("branchcfg", n),
				Name:                  "canary",
				Order:                 1,
				ApprovalOption:        app.InstallApprovalOptionPrompt,
				InstallApprovalOption: tt.groupOption,
			}
			persistHistoryRow(t, db, &group)

			groupID := group.ID
			switch {
			case tt.noGroupID:
				groupID = ""
			case tt.unknownGroup:
				groupID = historyID("missing", n)
			}

			wf, err := helpers.CreateWorkflow(ctx, install.ID, tt.workflowType, tt.metadata(groupID), false, nil)
			require.NoError(t, err)

			var persisted app.Workflow
			require.NoError(t, db.First(&persisted, "id = ?", wf.ID).Error)
			require.Equal(t, tt.wantOption, persisted.ApprovalOption)

			gotType, _ := persisted.Status.Metadata["approval_type"].(string)
			require.Equal(t, tt.wantType, gotType)
		})
	}
}
