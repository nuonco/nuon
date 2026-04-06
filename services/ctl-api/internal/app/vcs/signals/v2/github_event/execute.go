package githubevent

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/vcspush"
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

	// Only process push events for now.
	if event.EventType != "push" {
		l.Info(fmt.Sprintf("ignoring non-push event type: %s", event.EventType))
		return nil
	}

	// Parse the push payload for repo and branch.
	pushInfo, err := parsePushEvent(eventResp.Payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse push event payload: %v", err))
		return nil
	}

	l.Info(fmt.Sprintf("processing push event for repo=%s branch=%s install_id=%s",
		pushInfo.Repo, pushInfo.Branch, event.GithubInstallID))

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
			Repo:            pushInfo.Repo,
			Branch:          pushInfo.Branch,
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
				},
			})
			if err != nil {
				l.Error(fmt.Sprintf("failed to enqueue vcs-push signal for app branch %s: %v", match.AppBranchID, err))
				continue
			}

			l.Info(fmt.Sprintf("enqueued vcs-push signal for app branch %s", match.AppBranchID))
		}
	}

	return nil
}
