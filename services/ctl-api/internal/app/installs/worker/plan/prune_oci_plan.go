package plan

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

type CreatePruneOCIPlanRequest struct {
	InstallID        string
	DryRun           bool
	OwnedTags        []string
	OwnedTagPatterns []string
	KeepTags         []string
	GracePeriod      time.Duration
	MaxDeletes       int

	WorkflowID string
}

type CreatePruneOCIPlanResponse struct {
	Plan       *plantypes.PruneOCIPlan
	SkipReason string
}

// @temporal-gen-v2 workflow
// @execution-timeout 5m
// @task-timeout 1m
// @task-queue "api"
// @id-template {{.Req.WorkflowID}}
func CreatePruneOCIPlan(ctx workflow.Context, req *CreatePruneOCIPlanRequest) (*CreatePruneOCIPlanResponse, error) {
	p := Planner{}
	return p.createPruneOCIPlan(ctx, req)
}

func (p *Planner) createPruneOCIPlan(ctx workflow.Context, req *CreatePruneOCIPlanRequest) (*CreatePruneOCIPlanResponse, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	if install.SandboxMode.Bool {
		return &CreatePruneOCIPlanResponse{SkipReason: "install is in sandbox mode"}, nil
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}
	if stack.InstallStackOutputs.AWSStackOutputs == nil {
		return &CreatePruneOCIPlanResponse{SkipReason: "registry pruning only supports aws installs"}, nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	roleName, err := operationroles.MaintenanceRoleName(appCfg, installState)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get maintenance role")
	}
	roleSelection, err := operationroles.SelectDefaultRole(&operationroles.SelectionContext{
		Operation:               app.OperationDeploy,
		DefaultRole:             roleName,
		StackOutputs:            &stack.InstallStackOutputs,
		AppConfig:               appCfg,
		InstallState:            installState,
		LegacyAzureRoleFallback: p.legacyAzureRoleFallbackEnabled(ctx, l),
		LegacyGCPRoleFallback:   gcpLegacyRoleFallback,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to select maintenance role")
	}

	cloudAuth, err := p.getAuthForDeploy(ctx, roleSelection, stack, fmt.Sprintf("oci-prune-%s", install.ID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get auth for install registry")
	}

	stateData, err := installState.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "state data")
	}

	registry, err := p.RenderInstallRegistryRepository(l, &RenderInstallRegistryRepositoryInput{
		InstallDeploy: &app.InstallDeploy{InstallID: install.ID},
		Stack:         stack,
		StateData:     stateData,
		CloudAuth:     cloudAuth,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to render install registry")
	}

	clusterInfo, err := p.resolveKubernetesContextByName(ctx, "", appCfg, stack, installState, cloudAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve cluster")
	}
	if clusterInfo == nil {
		return &CreatePruneOCIPlanResponse{SkipReason: "install sandbox has no cluster to check for images in use"}, nil
	}

	return &CreatePruneOCIPlanResponse{
		Plan: &plantypes.PruneOCIPlan{
			Registry:           registry,
			DryRun:             req.DryRun,
			OwnedTags:          req.OwnedTags,
			OwnedTagPatterns:   req.OwnedTagPatterns,
			KeepTags:           req.KeepTags,
			GracePeriodSeconds: int64(req.GracePeriod / time.Second),
			MaxDeletes:         req.MaxDeletes,
			ClusterInfo:        clusterInfo,
		},
	}, nil
}
