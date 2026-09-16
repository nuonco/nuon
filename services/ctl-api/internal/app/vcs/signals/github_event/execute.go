package githubevent

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/syncinstalls"
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
	if pushInfo.Deleted {
		l.Info("ignoring deleted git ref")
		return nil
	}

	eventType := "push"
	headRef := ""
	if pushInfo.Tag != "" {
		eventType = "tag"
		headRef = pushInfo.Tag
	}

	l.Info(fmt.Sprintf("processing push event for repo=%s branch=%s tag=%s vcs_connection=%s",
		pushInfo.Repo, pushInfo.Branch, pushInfo.Tag, connEvent.VCSConnectionID))

	return s.fanOutToAppBranches(ctx, l, connEvent, fanOutRequest{
		Repo:         pushInfo.Repo,
		Branch:       pushInfo.Branch,
		PlanOnly:     false,
		EventType:    eventType,
		HeadSHA:      pushInfo.HeadSHA,
		HeadRef:      headRef,
		BaseSHA:      pushInfo.BeforeSHA,
		PusherEmails: pushInfo.PusherEmails,
		SenderLogin:  pushInfo.SenderLogin,
		ChangedFiles: pushInfo.ChangedFiles,
	})
}

func (s *Signal) handlePullRequestEvent(ctx workflow.Context, l *zap.Logger, connEvent *app.VCSConnectionEvent, event *app.GithubEvent, payload map[string]any) error {
	prInfo, err := parsePullRequestEvent(payload)
	if err != nil {
		l.Info(fmt.Sprintf("unable to parse pull_request event payload: %v", err))
		return nil
	}

	if prInfo.Action != "opened" && prInfo.Action != "synchronize" && prInfo.Action != "ready_for_review" {
		l.Info(fmt.Sprintf("ignoring pull_request action: %s", prInfo.Action))
		return nil
	}

	l.Info(fmt.Sprintf("processing pull_request event for repo=%s base=%s pr=%d head=%s ref=%s draft=%t vcs_connection=%s",
		prInfo.Repo, prInfo.BaseBranch, prInfo.PRNumber, prInfo.HeadSHA, prInfo.HeadRef, prInfo.Draft, connEvent.VCSConnectionID))

	return s.fanOutToAppBranches(ctx, l, connEvent, fanOutRequest{
		Repo:       prInfo.Repo,
		Branch:     prInfo.BaseBranch,
		PlanOnly:   true,
		EventType:  "pull_request",
		PRNumber:   &prInfo.PRNumber,
		HeadSHA:    prInfo.HeadSHA,
		HeadRef:    prInfo.HeadRef,
		BaseBranch: prInfo.BaseBranch,
		Draft:      prInfo.Draft,
	})
}

type fanOutRequest struct {
	Repo         string
	Branch       string
	PlanOnly     bool
	EventType    string
	PRNumber     *int
	HeadSHA      string
	HeadRef      string
	BaseBranch   string
	BaseSHA      string
	PusherEmails []string
	SenderLogin  string
	ChangedFiles []string
	Draft        bool
}

func (s *Signal) fanOutToAppBranches(ctx workflow.Context, l *zap.Logger, connEvent *app.VCSConnectionEvent, req fanOutRequest) error {
	matches, err := activities.AwaitFindMatchingAppBranches(ctx, activities.FindMatchingAppBranchesRequest{
		OrgID:  connEvent.OrgID,
		Repo:   req.Repo,
		Branch: req.Branch,
	})
	if err != nil {
		return errors.Wrap(err, "failed to find matching app branches")
	}

	for _, match := range matches {
		if !matchesRunConfig(match.RunConfig, req) {
			continue
		}
		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   match.AppBranchID,
			OwnerType: "app_branches",
			Signal: &vcspush.Signal{
				AppBranchID:         match.AppBranchID,
				AppBranchConfigID:   match.AppBranchConfigID,
				PlanOnly:            req.PlanOnly,
				EventType:           req.EventType,
				PRNumber:            req.PRNumber,
				HeadSHA:             req.HeadSHA,
				HeadRef:             req.HeadRef,
				BaseBranch:          req.BaseBranch,
				BaseSHA:             req.BaseSHA,
				ChangedFiles:        req.ChangedFiles,
				PusherEmails:        req.PusherEmails,
				SenderLogin:         req.SenderLogin,
				FallbackCreatedByID: connEvent.CreatedByID,
				Draft:               req.Draft,
			},
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to enqueue vcs-push signal for app branch %s: %v", match.AppBranchID, err))
			continue
		}
		l.Info(fmt.Sprintf("enqueued vcs-push signal for app branch %s (event_type=%s)", match.AppBranchID, req.EventType))
	}

	var installSyncMatches []activities.MatchingInstallSyncApp
	if req.Branch != "" {
		installSyncMatches, err = activities.AwaitFindMatchingInstallSyncApps(ctx, activities.FindMatchingInstallSyncAppsRequest{
			OrgID:  connEvent.OrgID,
			Repo:   req.Repo,
			Branch: req.Branch,
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to find matching install sync apps: %v", err))
		}
	}

	for _, match := range installSyncMatches {
		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   match.AppID,
			OwnerType: "apps",
			QueueName: appshelpers.AppInstallSyncsQueueName,
			Signal: &syncinstalls.Signal{
				AppID:               match.AppID,
				CommitSHA:           req.HeadSHA,
				TriggeredBy:         "vcs-push",
				FallbackCreatedByID: connEvent.CreatedByID,
			},
		})
		if err != nil {
			l.Error(fmt.Sprintf("failed to enqueue sync-installs signal for app %s: %v", match.AppID, err))
			continue
		}
		l.Info(fmt.Sprintf("enqueued sync-installs signal for app %s", match.AppID))
	}

	if len(matches) == 0 && len(installSyncMatches) == 0 {
		l.Info(fmt.Sprintf("no matching app branches or install sync apps for connection %s", connEvent.VCSConnectionID))
	}

	return nil
}

func matchesRunConfig(config app.AppBranchRunConfig, req fanOutRequest) bool {
	config.Normalize()
	switch req.EventType {
	case "pull_request":
		return config.Mode == app.AppBranchRunModePush
	case "tag":
		return config.Mode == app.AppBranchRunModeTagPrefix && strings.HasPrefix(req.HeadRef, config.TagPrefix)
	case "push":
		return config.Mode == app.AppBranchRunModePush || config.Mode == app.AppBranchRunModeGithubLabel
	default:
		return false
	}
}
