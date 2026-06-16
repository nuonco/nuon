package githubevent

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/vcspush"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	resp, err := activities.AwaitGetVCSConnectionEvent(ctx, activities.GetVCSConnectionEventRequest{
		VCSConnectionEventID: s.VCSConnectionEventID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get vcs connection event")
	}

	connEvent := resp.VCSConnectionEvent
	event := resp.GithubEvent

	l.Info(fmt.Sprintf("processing github event %s for vcs connection %s (event_type=%s)",
		event.ID, connEvent.VCSConnectionID, event.EventType))

	switch event.EventType {
	case "push":
		return s.handlePushEvent(ctx, l, connEvent, event, resp.Payload)
	case "pull_request":
		return s.handlePullRequestEvent(ctx, l, connEvent, event, resp.Payload)
	default:
		l.Info(fmt.Sprintf("ignoring event type: %s", event.EventType))
		return nil
	}
}

func (s *Signal) handlePushEvent(ctx workflow.Context, l *zap.Logger, connEvent *app.VCSConnectionEvent, event *app.GithubEvent, payload map[string]any) error {
	pushInfo, err := parsePushEvent(payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse push event payload: %v", err))
		return nil
	}

	l.Info(fmt.Sprintf("processing push event for repo=%s branch=%s vcs_connection=%s",
		pushInfo.Repo, pushInfo.Branch, connEvent.VCSConnectionID))

	return s.fanOutToAppBranches(ctx, l, connEvent, pushInfo.Repo, pushInfo.Branch, false, "push", nil, "", "")
}

func (s *Signal) handlePullRequestEvent(ctx workflow.Context, l *zap.Logger, connEvent *app.VCSConnectionEvent, event *app.GithubEvent, payload map[string]any) error {
	prInfo, err := parsePullRequestEvent(payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse pull_request event payload: %v", err))
		return nil
	}

	if prInfo.Action != "opened" && prInfo.Action != "synchronize" {
		l.Info(fmt.Sprintf("ignoring pull_request action: %s", prInfo.Action))
		return nil
	}

	l.Info(fmt.Sprintf("processing pull_request event for repo=%s base=%s pr=%d head=%s vcs_connection=%s",
		prInfo.Repo, prInfo.BaseBranch, prInfo.PRNumber, prInfo.HeadSHA, connEvent.VCSConnectionID))

	return s.fanOutToAppBranches(ctx, l, connEvent, prInfo.Repo, prInfo.BaseBranch, true, "pull_request", &prInfo.PRNumber, prInfo.HeadSHA, prInfo.BaseBranch)
}

func (s *Signal) fanOutToAppBranches(ctx workflow.Context, l *zap.Logger, connEvent *app.VCSConnectionEvent, repo, branch string, planOnly bool, eventType string, prNumber *int, headSHA, baseBranch string) error {
	matches, err := activities.AwaitFindMatchingAppBranches(ctx, activities.FindMatchingAppBranchesRequest{
		OrgID:  connEvent.OrgID,
		Repo:   repo,
		Branch: branch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to find matching app branches")
	}

	if len(matches) == 0 {
		l.Info(fmt.Sprintf("no matching app branches for connection %s", connEvent.VCSConnectionID))
		return nil
	}

	l.Info(fmt.Sprintf("found %d matching app branches for connection %s", len(matches), connEvent.VCSConnectionID))

	for _, match := range matches {
		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   match.AppBranchID,
			OwnerType: "app_branches",
			Signal: &vcspush.Signal{
				AppBranchID:       match.AppBranchID,
				AppBranchConfigID: match.AppBranchConfigID,
				PlanOnly:          planOnly,
				EventType:         eventType,
				PRNumber:          prNumber,
				HeadSHA:           headSHA,
				BaseBranch:        baseBranch,
			},
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to enqueue vcs-push signal for app branch %s: %v", match.AppBranchID, err))
			continue
		}

		l.Info(fmt.Sprintf("enqueued vcs-push signal for app branch %s (event_type=%s)", match.AppBranchID, eventType))
	}

	return nil
}
