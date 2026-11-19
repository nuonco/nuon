package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

var errAPIRequestRunbookExists = errors.New("API request runbook already exists")

func (h *Helpers) ReplayAPIRequestRunbook(ctx context.Context, install *app.Install, requestID, requestHash string) (*TriggerRunbookRunResponse, error) {
	var workflow app.Workflow
	err := h.db.WithContext(ctx).Where(&app.Workflow{
		OrgID: install.OrgID, OwnerID: install.ID, OwnerType: "installs", Type: app.WorkflowTypeRunbookRun,
	}, "OrgID", "OwnerID", "OwnerType", "Type").Where("request->>'request_id' = ?", requestID).First(&workflow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var current app.Install
	if err := h.db.WithContext(ctx).Where(&app.Install{ID: install.ID, AppID: install.AppID, OrgID: install.OrgID}, "ID", "AppID", "OrgID").First(&current).Error; err != nil {
		return nil, err
	}
	if err := request.Check(workflow.Request, requestHash, current.AppConfigID); err != nil {
		return nil, err
	}
	runID := generics.FromPtrStr(workflow.Metadata["install_runbook_run_id"])
	if runID == "" {
		return nil, fmt.Errorf("runbook request has no persisted run")
	}
	var run app.InstallRunbookRun
	if err := h.db.WithContext(ctx).Where(&app.InstallRunbookRun{
		ID: runID, OrgID: install.OrgID, InstallID: install.ID, InstallWorkflowID: &workflow.ID,
	}, "ID", "OrgID", "InstallID", "InstallWorkflowID").First(&run).Error; err != nil {
		return nil, err
	}
	var q app.Queue
	if err := h.db.WithContext(ctx).Where(&app.Queue{OrgID: &install.OrgID, OwnerID: install.ID, OwnerType: "installs", Name: "install-workflows"}, "OrgID", "OwnerID", "OwnerType", "Name").First(&q).Error; err != nil {
		return nil, err
	}
	dedupe := request.DedupeKey("runbook-run", requestID)
	resp, _ := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{QueueID: q.ID, Signal: executeflow.NewSignal(workflow.ID), OwnerID: workflow.ID, OwnerType: "install_workflows", DedupeKey: &dedupe})
	queueSignalID := ""
	if resp != nil {
		queueSignalID = resp.ID
	}
	return &TriggerRunbookRunResponse{Run: &run, Workflow: &workflow, QueueSignalID: queueSignalID}, nil
}
