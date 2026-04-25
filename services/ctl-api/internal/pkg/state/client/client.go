package client

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

const (
	workflowIDTemplate = "state-manager-%s"
	workflowNamespace  = "installs"
)

type Client struct {
	db      *gorm.DB
	cfg     *internal.Config
	tClient temporalclient.Client
	l       *zap.Logger
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Cfg     *internal.Config
	TClient temporalclient.Client
	L       *zap.Logger
}

func New(params Params) *Client {
	return &Client{
		db:      params.DB,
		cfg:     params.Cfg,
		tClient: params.TClient,
		l:       params.L,
	}
}

func workflowID(installID string) string {
	return fmt.Sprintf(workflowIDTemplate, installID)
}

// stateManagerStartOperation builds a WithStartWorkflowOperation for the state-manager workflow.
func (c *Client) stateManagerStartOperation(installID string) tclient.WithStartWorkflowOperation {
	req := statemanager.StateManagerRequest{
		InstallID: installID,
	}
	startOpts := tclient.StartWorkflowOptions{
		ID:        workflowID(installID),
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"type":       "state-manager",
			"install-id": installID,
		},
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}
	return c.tClient.NewWithStartWorkflowOperation(startOpts, "StateManager", req)
}
