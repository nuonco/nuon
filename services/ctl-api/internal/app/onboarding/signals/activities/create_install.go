package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/appconfigupdated"
	installscreated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/created"
	polldependencies "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/polldependencies"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateOnboardingInstallInput struct {
	OnboardingID string                           `json:"onboarding_id" validate:"required"`
	AppID        string                           `json:"app_id" validate:"required"`
	Name         string                           `json:"name" validate:"required"`
	AWSAccount   *CreateOnboardingInstallAWS      `json:"aws_account,omitempty"`
	AzureAccount *CreateOnboardingInstallAzure    `json:"azure_account,omitempty"`
	Inputs       map[string]*string               `json:"inputs,omitempty"`
	Config       *CreateOnboardingInstallConfig   `json:"install_config,omitempty"`
	Metadata     *CreateOnboardingInstallMetadata `json:"metadata,omitempty"`
}

type CreateOnboardingInstallAWS struct {
	Region string `json:"region"`
}

type CreateOnboardingInstallAzure struct {
	Location string `json:"location"`
}

type CreateOnboardingInstallConfig struct {
	ApprovalOption string `json:"approval_option"`
}

type CreateOnboardingInstallMetadata struct {
	ManagedBy string `json:"managed_by,omitempty"`
}

type CreateOnboardingInstallResponse struct {
	InstallID  string `json:"install_id"`
	WorkflowID string `json:"workflow_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) createOnboardingInstall(ctx context.Context, input *CreateOnboardingInstallInput) (*CreateOnboardingInstallResponse, error) {
	// Load org from onboarding — needed for GORM BeforeCreate hooks (OrgID)
	// and event loop startup (startEventLoop calls signal.GetOrg which needs org in context)
	var onboarding app.Onboarding
	if err := a.db.WithContext(ctx).First(&onboarding, "id = ?", input.OnboardingID).Error; err != nil {
		return nil, fmt.Errorf("unable to get onboarding: %w", err)
	}
	if onboarding.OrgID == nil || *onboarding.OrgID == "" {
		return nil, fmt.Errorf("onboarding has no org_id set")
	}
	var org app.Org
	if err := a.db.WithContext(ctx).First(&org, "id = ?", *onboarding.OrgID).Error; err != nil {
		return nil, fmt.Errorf("unable to get org: %w", err)
	}
	ctx = cctx.SetOrgContext(ctx, &org)

	sandboxMode := a.cfg.ForceOnboardingSandboxMode || org.SandboxMode || onboarding.InstallMode == app.OnboardingInstallModeSandbox
	installParams := &helpers.CreateInstallParams{
		Name:        input.Name,
		Inputs:      input.Inputs,
		SandboxMode: sandboxMode,
	}
	if input.AWSAccount != nil {
		installParams.AWSAccount = &struct {
			Region string `json:"region"`
		}{Region: input.AWSAccount.Region}
	}
	if input.AzureAccount != nil {
		installParams.AzureAccount = &struct {
			Location string `json:"location"`
		}{Location: input.AzureAccount.Location}
	}
	if input.Config != nil {
		installParams.InstallConfig = &helpers.CreateInstallConfigParams{
			ApprovalOption: app.InstallApprovalOption(input.Config.ApprovalOption),
		}
	}
	if input.Metadata != nil {
		installParams.Metadata = helpers.InstallMetadata{
			ManagedBy: input.Metadata.ManagedBy,
		}
	}

	install, err := a.installsHelpers.CreateInstall(ctx, input.AppID, installParams)
	if err != nil {
		return nil, fmt.Errorf("unable to create install: %w", err)
	}

	// Create provision workflow
	workflow, err := a.installsHelpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeProvision,
		map[string]string{},
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create provision workflow: %w", err)
	}

	useQueues, err := a.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		return nil, fmt.Errorf("checking features: %w", err)
	}

	if useQueues {
		signalsQueue, err := a.getInstallQueue(ctx, install.ID, helpers.InstallSignalsQueueName)
		if err != nil {
			return nil, err
		}
		workflowsQueue, err := a.getInstallQueue(ctx, install.ID, helpers.InstallWorkflowsQueueName)
		if err != nil {
			return nil, err
		}
		if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: signalsQueue.ID,
			Signal:  &installscreated.Signal{InstallID: install.ID},
		}); err != nil {
			return nil, fmt.Errorf("enqueue created signal: %w", err)
		}
		if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: signalsQueue.ID,
			Signal:  &polldependencies.Signal{InstallID: install.ID},
		}); err != nil {
			return nil, fmt.Errorf("enqueue polldependencies signal: %w", err)
		}
		if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   workflowsQueue.ID,
			OwnerID:   workflow.ID,
			OwnerType: "install_workflows",
			Signal:    &executeflow.Signal{WorkflowID: workflow.ID},
		}); err != nil {
			return nil, fmt.Errorf("enqueue executeflow signal: %w", err)
		}
		// reconcile cron/drift emitters from app config triggers
		if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: signalsQueue.ID,
			Signal:  &appconfigupdated.Signal{InstallID: install.ID},
		}); err != nil {
			return nil, fmt.Errorf("enqueue reconcile-emitters signal: %w", err)
		}
	} else {
		a.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationCreated,
		})
		a.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationPollDependencies,
		})
		a.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type: installsignals.OperationSyncActionWorkflowTriggers,
		})
		a.evClient.Send(ctx, install.ID, &installsignals.Signal{
			Type:              installsignals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}

	return &CreateOnboardingInstallResponse{
		InstallID:  install.ID,
		WorkflowID: workflow.ID,
	}, nil
}

func (a *Activities) getInstallQueue(ctx context.Context, installID, queueName string) (*app.Queue, error) {
	var queue app.Queue
	if res := a.db.WithContext(ctx).
		Where("owner_id = ? AND name = ?", installID, queueName).
		First(&queue); res.Error != nil {
		return nil, fmt.Errorf("unable to get install queue %s: %w", queueName, res.Error)
	}
	return &queue, nil
}
