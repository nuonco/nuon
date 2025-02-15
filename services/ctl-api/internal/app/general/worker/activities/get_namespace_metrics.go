package activities

import (
	"context"
	"regexp"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/api/workflowservice/v1"
	"gorm.io/gorm"
)

var eventLoopRegex = regexp.MustCompile(`^event-loop-[a-zA-Z0-9_-]{26}$`)

type GetNamespaceMetricsRequest struct {
	Name string `validate:"required"`
}

type NamespaceMetrics struct {
	EventLoops         int64
	AllWorkflows       int64
	ExpectedEventLoops int64
}

// @temporal-gen activity
// @by-id Name
// @schedule-to-close-timeout 120s
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

	var expectedEventLoops int64
	switch name {
	case "orgs":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.Org](ctx, name, a.db)
	case "runners":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.Runner](ctx, name, a.db)
	case "components":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.Component](ctx, name, a.db)
	case "apps":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.App](ctx, name, a.db)
	case "installs":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.Install](ctx, name, a.db)
	case "releases":
		expectedEventLoops, err = getNamespaceExpectedEventLoops[app.ComponentRelease](ctx, name, a.db)
	default:
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get expected event loops")
	}

	metrics.ExpectedEventLoops = expectedEventLoops

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

		for _, wf := range wfs.Executions {
			metrics.AllWorkflows += 1

			if eventLoopRegex.MatchString(wf.Execution.WorkflowId) {
				metrics.EventLoops += 1
			}
		}

		token = wfs.NextPageToken
		if len(token) < 1 {
			break
		}
	}

	return metrics, nil
}

func getNamespaceExpectedEventLoops[T any](ctx context.Context, ns string, db *gorm.DB) (int64, error) {
	var obj T
	var count int64

	tx := db.
		WithContext(ctx).
		Model(&obj)

	if ns != "orgs" {
		tx = tx.Joins("JOIN orgs ON orgs.id = org_id").
			Where("orgs.org_type IN (?)", []string{"default", "real", "sandbox"})
	}

	if res := tx.Count(&count); res.Error != nil {
		return 0, errors.Wrap(res.Error, "unable to get count")
	}

	return count, nil
}
