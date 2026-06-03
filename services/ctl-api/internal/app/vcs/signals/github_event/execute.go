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

	// Fetch the GithubEvent with its parsed payload.
	eventResp, err := activities.AwaitGetGithubEvent(ctx, activities.GetGithubEventRequest{
		GithubEventID: s.GithubEventID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get github event")
	}

	event := eventResp.GithubEvent

	switch event.EventType {
	case "push":
		return s.handlePushEvent(ctx, l, event, eventResp.Payload)
	case "pull_request":
		return s.handlePullRequestEvent(ctx, l, event, eventResp.Payload)
	default:
		l.Info(fmt.Sprintf("ignoring event type: %s", event.EventType))
		return nil
	}
}

func (s *Signal) handlePushEvent(ctx workflow.Context, l *zap.Logger, event *app.GithubEvent, payload map[string]any) error {
	pushInfo, err := parsePushEvent(payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse push event payload: %v", err))
		return nil
	}

	l.Info(fmt.Sprintf("processing push event for repo=%s branch=%s install_id=%s",
		pushInfo.Repo, pushInfo.Branch, event.GithubInstallID))

	return s.fanOutToAppBranches(ctx, l, event, pushInfo.Repo, pushInfo.Branch, false, "push", nil, "", "")
}

func (s *Signal) handlePullRequestEvent(ctx workflow.Context, l *zap.Logger, event *app.GithubEvent, payload map[string]any) error {
	prInfo, err := parsePullRequestEvent(payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse pull_request event payload: %v", err))
		return nil
	}

	// Only process opened and synchronize actions
	if prInfo.Action != "opened" && prInfo.Action != "synchronize" {
		l.Info(fmt.Sprintf("ignoring pull_request action: %s", prInfo.Action))
		return nil
	}

	l.Info(fmt.Sprintf("processing pull_request event for repo=%s base=%s pr=%d head=%s",
		prInfo.Repo, prInfo.BaseBranch, prInfo.PRNumber, prInfo.HeadSHA))

	// Match against the base branch (the branch the PR targets)
	return s.fanOutToAppBranches(ctx, l, event, prInfo.Repo, prInfo.BaseBranch, true, "pull_request", &prInfo.PRNumber, prInfo.HeadSHA, prInfo.BaseBranch)
}

func (s *Signal) fanOutToAppBranches(ctx workflow.Context, l *zap.Logger, event *app.GithubEvent, repo, branch string, planOnly bool, eventType string, prNumber *int, headSHA, baseBranch string) error {
	// Find all VCS connections for this GitHub installation (across orgs).
	connsResp, err := activities.AwaitFindVCSConnectionsByInstallID(ctx, activities.FindVCSConnectionsByInstallIDRequest{
		GithubInstallID: event.GithubInstallID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to find vcs connections")
	}

	if len(connsResp.VCSConnections) == 0 {
		l.Info("no vcs connections found for github install id")
		return nil
	}

	l.Info(fmt.Sprintf("found %d vcs connections for install_id=%s", len(connsResp.VCSConnections), event.GithubInstallID))

	// Fan out: for each connection, create a VCSConnectionEvent and find matching app branches.
	for _, conn := range connsResp.VCSConnections {
		// Create the VCSConnectionEvent linking this github event to the connection.
		_, err := activities.AwaitCreateVCSConnectionEvent(ctx, activities.CreateVCSConnectionEventRequest{
			VCSConnectionID: conn.ID,
			GithubEventID:   s.GithubEventID,
			OrgID:           conn.OrgID,
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to create vcs connection event for connection %s: %v", conn.ID, err))
			continue
		}

		// Find matching app branches for this connection + repo + branch.
		matches, err := activities.AwaitFindMatchingAppBranches(ctx, activities.FindMatchingAppBranchesRequest{
			VCSConnectionID: conn.ID,
			Repo:            repo,
			Branch:          branch,
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to find matching app branches for connection %s: %v", conn.ID, err))
			continue
		}

		if len(matches) == 0 {
			l.Info(fmt.Sprintf("no matching app branches for connection %s", conn.ID))
			continue
		}

		l.Info(fmt.Sprintf("found %d matching app branches for connection %s", len(matches), conn.ID))

		// Fan out vcs-push signals to each matching app branch.
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
	}

	return nil
}
