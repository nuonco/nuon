package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/api/workflowservice/v1"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
)

type TerminateNamespaceWorkflowsRequest struct {
	Namespace string `json:"namespace" validate:"required"`
}

// @temporal-gen activity
// @by-id Namespace
// @schedule-to-close-timeout 120s
// @start-to-close-timeout 120s
func (a *Activities) TerminateNamespaceWorkflows(ctx context.Context, req TerminateNamespaceWorkflowsRequest) (int, error) {
	client, err := a.tClient.GetNamespaceClient(req.Namespace)
	if err != nil {
		return 0, errors.Wrap(err, "unable to get client")
	}

	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("namespace", req.Namespace))

	cnt := 0
	var token []byte
	for {
		wfs, err := client.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
			MaximumPageSize: int32(200),
			Namespace:       req.Namespace,
			NextPageToken:   token,
		})
		if err != nil {
			return cnt, errors.Wrap(err, "unable to get namespace workflows")
		}

		for _, wf := range wfs.Executions {
			if !eventLoopRegex.MatchString(wf.Execution.WorkflowId) {
				continue
			}

			cnt += 1
			l.Info("terminating workflow",
				zap.String("workflow-id", wf.Execution.WorkflowId),
			)

			if err := client.TerminateWorkflow(ctx, wf.Execution.WorkflowId, "", "terminating from general event loop"); err != nil {
				l.Error("unable to terminate workflow", zap.Error(err))
			}
		}

		token = wfs.NextPageToken
		if len(token) < 1 {
			break
		}
	}

	return cnt, nil
}
