package temporal

import (
	"context"
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

func (r *repo) TriggerInstallProvision(ctx context.Context, req *installsv1.ProvisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": "nuonctl",
		},
	}

	_, err := r.Client.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to provision install: %w", err)
	}

	return nil
}

func (r *repo) TriggerInstallDeprovision(ctx context.Context, req *installsv1.DeprovisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"org-id":     req.OrgId,
			"app-id":     req.AppId,
			"install-id": req.InstallId,
			"started-by": "nuonctl",
		},
	}

	_, err := r.Client.ExecuteWorkflow(ctx, opts, "Deprovision", req)
	if err != nil {
		return fmt.Errorf("unable to deprovision install: %w", err)
	}

	return nil
}
