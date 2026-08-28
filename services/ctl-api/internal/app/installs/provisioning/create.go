package provisioning

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appconfigupdated"
	installscreated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/polldependencies"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func Create(ctx context.Context, installsHelpers *helpers.Helpers, db *gorm.DB, queueClient *queueclient.Client, appID string, params *helpers.CreateInstallParams) (*app.Install, *app.Workflow, error) {
	install, workflow, err := Prepare(ctx, installsHelpers, db, appID, params)
	if err != nil {
		return nil, nil, err
	}
	if err := Enqueue(ctx, db, queueClient, install, workflow); err != nil {
		return nil, nil, err
	}
	return install, workflow, nil
}

func Prepare(ctx context.Context, installsHelpers *helpers.Helpers, db *gorm.DB, appID string, params *helpers.CreateInstallParams) (*app.Install, *app.Workflow, error) {
	install, err := installsHelpers.CreateInstall(ctx, appID, params)
	if err != nil {
		return nil, nil, err
	}
	workflow, err := installsHelpers.CreateWorkflow(ctx, install.ID, app.WorkflowTypeProvision, map[string]string{}, false)
	if err != nil {
		return nil, nil, fmt.Errorf("create provision workflow: %w", err)
	}
	lifecycle := lifecyclephase.New(lifecyclephase.Provisioning, "Setting up runner and sandbox resources")
	if err := db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Update("lifecycle_phase", lifecycle).Error; err != nil {
		return nil, nil, fmt.Errorf("set install lifecycle phase: %w", err)
	}
	install.WorkflowID = &workflow.ID
	return install, workflow, nil
}

func Enqueue(ctx context.Context, db *gorm.DB, queueClient *queueclient.Client, install *app.Install, workflow *app.Workflow) error {
	var queues []app.Queue
	if err := db.WithContext(ctx).Where(app.Queue{OwnerID: install.ID, OwnerType: "installs"}).Find(&queues).Error; err != nil {
		return fmt.Errorf("load install queues: %w", err)
	}
	queueIDs := make(map[string]string, len(queues))
	for _, queue := range queues {
		queueIDs[queue.Name] = queue.ID
	}
	if queueIDs[helpers.InstallSignalsQueueName] == "" || queueIDs[helpers.InstallWorkflowsQueueName] == "" {
		return fmt.Errorf("install queues are incomplete")
	}
	requests := []*queueclient.EnqueueSignalRequest{
		{QueueID: queueIDs[helpers.InstallSignalsQueueName], Signal: &installscreated.Signal{InstallID: install.ID}},
		{QueueID: queueIDs[helpers.InstallSignalsQueueName], Signal: &polldependencies.Signal{InstallID: install.ID}},
		{QueueID: queueIDs[helpers.InstallWorkflowsQueueName], Signal: &executeflow.Signal{WorkflowID: workflow.ID}, OwnerID: workflow.ID, OwnerType: "install_workflows"},
		{QueueID: queueIDs[helpers.InstallSignalsQueueName], Signal: &appconfigupdated.Signal{InstallID: install.ID}},
	}
	for _, request := range requests {
		if _, err := queueClient.EnqueueSignal(ctx, request); err != nil {
			return fmt.Errorf("enqueue install provisioning signal: %w", err)
		}
	}
	return nil
}
