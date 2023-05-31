package workflows

import (
	"context"
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	tclient "go.temporal.io/sdk/client"
)

func (wfClient *workflowsClient) TriggerInstallProvision(ctx context.Context, req *installsv1.ProvisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": "nuonctl",
		},
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to provision install: %w", err)
	}

	return nil
}

func (wfClient *workflowsClient) TriggerInstallDeprovision(ctx context.Context, req *installsv1.DeprovisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": "nuonctl",
		},
	}

	_, err := wfClient.TemporalClient.ExecuteWorkflowInNamespace(ctx, "installs", opts, "Deprovision", req)
	if err != nil {
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	return nil
}
