package activities

import (
	"context"
	"fmt"
	"strconv"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	appsigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	componentsigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	installsigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	orgsigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	releasesigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/signals"

	enumsv1 "go.temporal.io/api/enums/v1"
)

/*

The "parent" that fires off the "child" functions. Queries for namespaces and
does the ranging to invoke sub function.

NOTE(fd):
- The child/paginated work is currently invoked directly but we'll want to move
  it into its own workflow
- We make use of temporal queries. These requires advanced analytics.
  > docs: https://docs.temporal.io/visibility#supported-operators

*/

const DefaultPageSize int64 = 100

type EnsureEventLoopsRequest struct{}
type EnsureEventLoopsResponse struct {
	Namespace string
	RowCount  int64
}

type EnsureEventLoopsPageRequest struct {
	Namespace string
	Limit     int
	Offset    int
}

func namespacesToReconcile() []string {
	return []string{"apps", "orgs", "installs", "components", "releases"}
}

// @temporal-gen activity
// @schedule-to-close-timeout 60s
func (a *Activities) EnsureEventLoop(ctx context.Context, req EnsureEventLoopsRequest) ([]EnsureEventLoopsResponse, error) {
	// this Method fetches the row count for all the db tables whose rows' event loops we're going to check

	// prepare response
	var response []EnsureEventLoopsResponse

	// iterate through namespaces
	namespaces := namespacesToReconcile()
	for _, ns := range namespaces {
		var rowCount int64

		switch ns {
		case "apps":
			a.db.WithContext(ctx).Model(&app.App{}).Count(&rowCount)
		case "orgs":
			a.db.WithContext(ctx).Model(&app.Org{}).Count(&rowCount)
		case "installs":
			a.db.WithContext(ctx).Model(&app.Install{}).Count(&rowCount)
		case "components":
			a.db.WithContext(ctx).Model(&app.Component{}).Count(&rowCount)
		case "releases":
			a.db.WithContext(ctx).Model(&app.ComponentRelease{}).Count(&rowCount)
		default:
			rowCount = 0
		}

		response = append(response, EnsureEventLoopsResponse{
			Namespace: ns,
			RowCount:  rowCount,
		})
	}

	return response, nil
}

/*

We abstracted out getting the workflow ids into a function.

This is not an Activity, per se, but we keep the pattern so we can use the context.
We _could_ break this work down into Parent > Child > SubChild and pass the ids back that way, but that seems extra.

*/

func (a *Activities) GetWorkflowIds(ctx context.Context, req EnsureEventLoopsPageRequest) ([]string, error) {
	var objIds []string

	switch req.Namespace {
	case "orgs":
		a.db.WithContext(ctx).Offset(req.Offset).Limit(req.Limit).Model(&app.Org{}).Pluck("id", &objIds)
	case "installs":
		a.db.WithContext(ctx).Offset(req.Offset).Limit(req.Limit).Model(&app.Install{}).Pluck("id", &objIds)
	case "components":
		a.db.WithContext(ctx).Offset(req.Offset).Limit(req.Limit).Model(&app.Component{}).Pluck("id", &objIds)
	case "releases":
		a.db.WithContext(ctx).Offset(req.Offset).Limit(req.Limit).Model(&app.ComponentRelease{}).Pluck("id", &objIds)
	}
	return objIds, nil
}

/*

The "child" that handles paginating through the namespace workflows.

This function's request holds a namespace, limit, and offset which we use to get a "page" of results from the db.
With this result in hand, we then fetch the object IDs and query the status of each workflow.

*/

// @temporal-gen activity
// @schedule-to-close-timeout 60s
func (a *Activities) EnsureEventLoopPage(ctx context.Context, req EnsureEventLoopsPageRequest) (int, error) {
	ids, err := a.GetWorkflowIds(ctx, req)
	if err != nil {
		a.logger.Error(fmt.Sprintf("%v", err), "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))
		return 0, err
	} else if len(ids) == 0 {
		a.logger.Error("No IDs, exiting early.", "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))
		return 0, nil
	}
	a.logger.Debug(fmt.Sprintf("IdCount=%d", len(ids)), "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))

	// check if the workflows are running
	// 1. get a count of total workflows (empty query)
	// 2. get the worfklow status (enum)
	// NOTE(fd): the ones we care about are event-loop-<ID>.
	for _, id := range ids {

		elId := fmt.Sprintf("event-loop-%s", id)

		wfCount, err := a.evClient.GetWorkflowCount(ctx, req.Namespace, elId)
		if err != nil {
			a.logger.Error(fmt.Sprintf("%v", err), "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))
			return 0, err
		}

		status, err := a.evClient.GetWorkflowStatus(ctx, req.Namespace, elId)
		if err != nil {
			a.logger.Error(fmt.Sprintf("%v", err), "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))
			return 0, err
		}

		a.logger.Debug("Retrieved Status", "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset), "Status", status.String())
		if status == enumsv1.WORKFLOW_EXECUTION_STATUS_RUNNING {
			// happy path: do nothing
		} else if status == enumsv1.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED {
			// NOTE(fd): this sends a restart signal to event loops that are not running including those that have not started
			a.mw.Incr("event_loop.restart", []string{"id", id})
			switch req.Namespace {
			case "apps":
				a.evClient.Send(ctx, id, &appsigs.Signal{Type: appsigs.OperationRestart})
			case "orgs":
				a.evClient.Send(ctx, id, &orgsigs.Signal{Type: orgsigs.OperationRestart})
			case "installs":
				a.evClient.Send(ctx, id, &installsigs.Signal{Type: installsigs.OperationRestart})
			case "components":
				a.evClient.Send(ctx, id, &componentsigs.Signal{Type: componentsigs.OperationRestart})
			case "releases":
				a.evClient.Send(ctx, id, &releasesigs.Signal{Type: releasesigs.OperationRestart})
			}
		}
		a.logger.Debug(fmt.Sprintf("status: %s", status.String()), "Activity", "EnsureEventLoopPage", "Namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "Offset", strconv.Itoa(req.Offset))

		// emit metrics
		a.mw.Incr("event_loop.status", []string{"status", status.String(), "id", id, "namespace", req.Namespace, "Limit", strconv.Itoa(req.Limit), "offset", strconv.Itoa(req.Offset)})
		a.mw.Gauge("event_loop.executions", float64(wfCount), []string{"status", status.String(), "id", id, "namespace", req.Namespace})
	}

	return 0, nil
}

// @temporal-gen activity
// @schedule-to-close-timeout 60s
func (a *Activities) SendReconcileSignal(ctx context.Context, req EnsureEventLoopsRequest) error {
	a.evClient.Send(ctx, signals.EventLoop, &signals.Signal{Type: signals.OperationReconcile})
	return nil
}
