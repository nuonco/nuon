package helpers

import (
	"context"
	"fmt"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

// componentHealthSessionName keeps health's assume-role sessions separable from
// deploys in the customer's cloud audit log.
const componentHealthSessionName = "nuon-component-health"

// ComponentHealthClusterAccess describes the access health will read through,
// and the roles it could have used instead.
type ComponentHealthClusterAccess struct {
	ClusterInfo *kube.ClusterInfo
	RoleName    string
	Roles       map[string]string
}

// ResolveComponentHealthClusterAccess builds the cluster access the health
// engine needs from the same install outputs a deploy plan templates against,
// so an install can be watched without waiting for its next deploy.
//
// roleName picks the identity to read through; empty means the maintenance
// role, matching every other day-2 operation (drift, action runs).
//
// Returns nil when the install has no cluster to watch — a Lambda or ECS
// sandbox emits no cluster output, and that is an answer rather than a failure.
func (h *Helpers) ResolveComponentHealthClusterAccess(ctx context.Context, installID, roleName string) (*ComponentHealthClusterAccess, error) {
	in, err := h.componentHealthClusterInputs(ctx, installID)
	if err != nil {
		return nil, err
	}
	if !sandboxEmitsClusterOutputs(in.stateData) {
		return nil, nil
	}

	roles, err := operationroles.AvailableRoles(in.appCfg, &in.stack.InstallStackOutputs, in.installState)
	if err != nil {
		return nil, fmt.Errorf("unable to list available roles: %w", err)
	}

	if roleName == "" {
		roleName, err = operationroles.MaintenanceRoleName(in.appCfg, in.installState)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve default role: %w", err)
		}
	}
	roleID, ok := roles[roleName]
	if !ok {
		return nil, fmt.Errorf("role %q is not available on this install", roleName)
	}

	clusterInfo := componentHealthClusterInfo(&in.stack.InstallStackOutputs, roleID)
	if clusterInfo == nil {
		return nil, nil
	}

	if err := render.RenderStruct(clusterInfo, in.stateData); err != nil {
		return nil, fmt.Errorf("unable to render cluster info: %w", err)
	}

	// An unresolved template renders empty; half a cluster is unusable.
	if clusterInfo.ID == "" || clusterInfo.Endpoint == "" || clusterInfo.CAData == "" {
		return nil, nil
	}

	return &ComponentHealthClusterAccess{
		ClusterInfo: clusterInfo,
		RoleName:    roleName,
		Roles:       roles,
	}, nil
}

type componentHealthClusterInput struct {
	stack        *app.InstallStack
	installState *state.State
	stateData    map[string]any
	appCfg       *app.AppConfig
}

func (h *Helpers) componentHealthClusterInputs(ctx context.Context, installID string) (*componentHealthClusterInput, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "app_config_id").
		Where(app.Install{ID: installID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	stack, err := h.getInstallStack(ctx, installID)
	if err != nil {
		return nil, err
	}

	installState, err := h.GetInstallState(ctx, installID, false, true)
	if err != nil {
		return nil, fmt.Errorf("unable to get install state: %w", err)
	}
	stateData, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to build install state map: %w", err)
	}

	appCfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, true)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}

	return &componentHealthClusterInput{
		stack:        stack,
		installState: installState,
		stateData:    stateData,
		appCfg:       appCfg,
	}, nil
}

// componentHealthClusterInfo mirrors the deploy planner's cluster templating,
// pinned to the sandbox default context and the caller's chosen identity.
func componentHealthClusterInfo(outputs *app.InstallStackOutputs, roleID string) *kube.ClusterInfo {
	const clusterPath = ".nuon.sandbox.outputs.cluster"

	switch {
	case outputs.AWSStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:       fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint: fmt.Sprintf("{{%s.endpoint}}", clusterPath),
			CAData:   fmt.Sprintf("{{%s.certificate_authority_data}}", clusterPath),
			AWSAuth: &awscredentials.Config{
				Region: outputs.AWSStackOutputs.Region,
				AssumeRole: &awscredentials.AssumeRoleConfig{
					SessionName: componentHealthSessionName,
					RoleARN:     roleID,
				},
			},
		}

	case outputs.AzureStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:       fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint: fmt.Sprintf("{{%s.host}}", clusterPath),
			CAData:   fmt.Sprintf("{{%s.cluster_ca_certificate}}", clusterPath),
			AzureAuth: &azurecredentials.Config{
				ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
					SubscriptionID:       outputs.AzureStackOutputs.SubscriptionID,
					SubscriptionTenantID: outputs.AzureStackOutputs.SubscriptionTenantID,
				},
				ManagedIdentityClientID: roleID,
			},
		}

	case outputs.GCPStackOutputs != nil:
		return &kube.ClusterInfo{
			ID:       fmt.Sprintf("{{%s.name}}", clusterPath),
			Endpoint: fmt.Sprintf("{{%s.endpoint}}", clusterPath),
			CAData:   fmt.Sprintf("{{%s.certificate_authority_data}}", clusterPath),
			GCPAuth: &gcpcredentials.Config{
				ProjectID:                 outputs.GCPStackOutputs.ProjectID,
				Region:                    outputs.GCPStackOutputs.Region,
				ImpersonateServiceAccount: roleID,
			},
		}
	}

	return nil
}

// sandboxEmitsClusterOutputs reports whether the sandbox emitted a `cluster`
// output, the same signal deploy planning uses to pick the default context.
func sandboxEmitsClusterOutputs(stateData map[string]any) bool {
	sandbox, ok := stateData["sandbox"].(map[string]any)
	if !ok {
		return false
	}
	outputs, ok := sandbox["outputs"].(map[string]any)
	if !ok {
		return false
	}
	cluster, ok := outputs["cluster"]
	return ok && cluster != nil
}
