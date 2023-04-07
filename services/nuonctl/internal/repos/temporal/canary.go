package temporal

import (
	"context"
	"fmt"

	canaryv1 "github.com/powertoolsdev/mono/pkg/types/workflows/canary/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	tclient "go.temporal.io/sdk/client"
)

func (r *repo) TriggerCanaryProvision(ctx context.Context, req *canaryv1.ProvisionRequest) error {
	opts := tclient.StartWorkflowOptions{
		TaskQueue: workflows.DefaultTaskQueue,
		Memo: map[string]interface{}{
			"canary-id":  req.CanaryId,
			"started-by": "nuonctl",
		},
	}

	_, err := r.Client.ExecuteWorkflow(ctx, opts, "Provision", req)
	if err != nil {
		return fmt.Errorf("unable to provision install: %w", err)
	}

	return nil
}
