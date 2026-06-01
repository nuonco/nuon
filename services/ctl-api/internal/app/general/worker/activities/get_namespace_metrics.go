package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.temporal.io/api/workflowservice/v1"
)

type GetNamespaceMetricsRequest struct {
	Name string `validate:"required"`
}

type NamespaceMetrics struct {
	AllWorkflows int64
}

// @temporal-gen-v2 activity
// @by-field Name
// @schedule-to-close-timeout 120s
// @start-to-close-timeout 120s
func (a *Activities) GetNamespaceMetrics(ctx context.Context, req GetNamespaceMetricsRequest) (*NamespaceMetrics, error) {
	metrics, err := a.getNamespaceMetrics(ctx, req.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get namespace metrics")
	}

	return metrics, nil
}

func (a *Activities) getNamespaceMetrics(ctx context.Context, name string) (*NamespaceMetrics, error) {
	client, err := a.tClient.GetNamespaceClient(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get client")
	}

	metrics := &NamespaceMetrics{}

	var token []byte
	for {
		wfs, err := client.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
			MaximumPageSize: int32(200),
			Namespace:       name,
			NextPageToken:   token,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get namespace workflows")
		}

		for range wfs.Executions {
			metrics.AllWorkflows += 1
		}

		token = wfs.NextPageToken
		if len(token) < 1 {
			break
		}
	}

	return metrics, nil
}
