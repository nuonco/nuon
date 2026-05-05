package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

// SlackParams declares the dependencies for the Slack signal lifecycle hook.
// All fields are optional (mirroring webhook.go's defensive constructor) so
// the hook can be wired into FX even when Slack isn't configured locally.
type SlackParams struct {
	fx.In

	Cfg         *internal.Config    `optional:"true"`
	L           *zap.Logger         `optional:"true"`
	DB          *gorm.DB            `name:"psql" optional:"true"`
	SlackClient *slackclient.Client `optional:"true"`
}

// SlackSignalLifecycleHook fans out workflow / step / approval lifecycle
// events to all active SlackChannelSubscriptions for the event's org.
//
// Routing invariant (mirrors the model docs): a message lands in workspace T
// for org O iff installation T is active, org_link (T, O) is verified, and a
// channel sub (T, channel, O) is active. The hook resolves all three via
// per-event GORM lookups and posts via the handwritten Slack client.
//
// Threading: per-(team, channel, workflow) anchor rows in slack_thread_anchors
// drive a parent post + threaded children pattern. The first event posts a
// parent, persists its ts, and threads itself under that parent. Subsequent
// events thread under the cached parent and best-effort edit the parent's
// rollup. Nested action_workflow_run sub-workflows are consolidated under the
// launching deploy step's workflow via the parent lookup.
//
// Most enrichment helpers (enrichStep, lookupParent, buildContextLinks,
// lookupDeployTargetMeta, lookupSandboxRunTargetMeta, lookupApprovalResponse)
// live on WebhookSignalLifecycleHook in webhook.go and are reused via a
// lightweight delegate so both hooks share one source of truth for payload
// shape.
type SlackSignalLifecycleHook struct {
	l           *zap.Logger
	db          *gorm.DB
	slackClient *slackclient.Client
	appURL      string
	enricher    *WebhookSignalLifecycleHook
}

var _ signal.SignalLifecycleHook = (*SlackSignalLifecycleHook)(nil)

// NewSlackSignalLifecycleHook constructs the Slack lifecycle hook. Returns a
// non-nil hook even when dependencies are missing — Supports() short-circuits
// at runtime so the dispatcher cost stays cheap when Slack isn't configured.
func NewSlackSignalLifecycleHook(params SlackParams) *SlackSignalLifecycleHook {
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}

	appURL := ""
	publicAPIURL := ""
	if params.Cfg != nil {
		appURL = strings.TrimSpace(params.Cfg.AppURL)
		publicAPIURL = strings.TrimSpace(params.Cfg.PublicAPIURL)
	}

	// Reuse webhook.go's enrichment pipeline. Building a private instance
	// (rather than depending on the FX-wired hook) keeps Slack and webhook
	// independently constructible and avoids accidental cycles in the
	// dependency graph. The enricher only ever reads from the DB.
	enricher := &WebhookSignalLifecycleHook{
		l:            logger,
		db:           params.DB,
		appURL:       appURL,
		publicAPIURL: publicAPIURL,
	}

	return &SlackSignalLifecycleHook{
		l:           logger,
		db:          params.DB,
		slackClient: params.SlackClient,
		appURL:      appURL,
		enricher:    enricher,
	}
}

func (h *SlackSignalLifecycleHook) Name() string {
	return "workflow_lifecycle_slack"
}

// Supports limits this hook to the public lifecycle primitives (matches the
// webhook hook's filter) and short-circuits when Slack isn't wired so the
// dispatcher doesn't pay the per-event cost.
func (h *SlackSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.slackClient == nil || h.db == nil {
		return false
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow,
		signalTypeExecuteWorkflowStep,
		signalTypeWorkflowStepApprovalRequest,
		signalTypeWorkflowStepApprovalResponse,
		signalTypeDriftDetected:
		return true
	default:
		return false
	}
}

func (h *SlackSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	// Only emit *.started events on the execute phase.
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	// Approval signals don't have a meaningful "started" semantic — see the
	// matching comment in webhook.go. Drift-detected is a single-shot
	// notification carrier (its Execute is a no-op) so a "started" emission
	// would just produce a duplicate message right before the real one.
	if isApprovalSignalType(event.SignalType) || event.SignalType == signalTypeDriftDetected {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish workflow lifecycle slack message",
			zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func (h *SlackSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}

	h.l.Debug("workflow lifecycle slack after-phase",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("status", string(outcome.Status)),
	)

	return h.publish(ctx, event, &outcome)
}

// publish renders the slack messages for the event and dispatches to all
// eligible channel subscriptions. Delivery errors are aggregated
// (errors.Join) so a single failing workspace doesn't swallow others.
func (h *SlackSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	if event.OrgID == "" || event.WorkflowID == "" {
		return nil
	}

	// Reuse webhook.go's payload builder so the renderer sees exactly the
	// same enriched shape webhook consumers see (plus the same hidden-step
	// suppression rules).
	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	rendered := buildRenderEvent(data)

	// Resolve verified org-links for this org, then the matching active
	// installations keyed by team_id.
	var links []app.SlackOrgLink
	if err := h.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			OrgID:  event.OrgID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		Find(&links).Error; err != nil {
		return fmt.Errorf("unable to list slack org links for slack lifecycle: %w", err)
	}
	if len(links) == 0 {
		return nil
	}

	teamIDs := make([]string, 0, len(links))
	for _, link := range links {
		teamIDs = append(teamIDs, link.TeamID)
	}

	var installations []app.SlackInstallation
	if err := h.db.WithContext(ctx).
		Where("team_id IN ? AND status = ?", teamIDs, app.SlackInstallationStatusActive).
		Find(&installations).Error; err != nil {
		return fmt.Errorf("unable to list slack installations for slack lifecycle: %w", err)
	}

	installByTeam := make(map[string]*app.SlackInstallation, len(installations))
	for i := range installations {
		installByTeam[installations[i].TeamID] = &installations[i]
	}

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("org_id", event.OrgID),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("anchor_workflow_id", anchorWorkflowID(data)),
	)

	var sendErrs []error
	for _, link := range links {
		install, ok := installByTeam[link.TeamID]
		if !ok {
			continue
		}

		var subs []app.SlackChannelSubscription
		if err := h.db.WithContext(ctx).
			Where(app.SlackChannelSubscription{
				OrgLinkID: link.ID,
				OrgID:     event.OrgID,
			}).
			Find(&subs).Error; err != nil {
			logger.Warn("failed to list channel subscriptions",
				zap.String("team_id", link.TeamID), zap.Error(err))
			sendErrs = append(sendErrs, err)
			continue
		}

		for _, sub := range subs {
			if !interests.Matches(event, outcome, h.db, sub.Interests) {
				continue
			}

			// Drift-detected events bypass the parent-anchor / threaded-reply
			// machinery: they are the only meaningful signal subscribers get
			// from a drift scan (the surrounding drift_run /
			// drift_run_reprovision_sandbox lifecycle events are suppressed
			// in interests.Matches), so each detection is its own top-level
			// message linked directly to the affected component or sandbox.
			var err error
			if event.SignalType == signalTypeDriftDetected {
				err = h.postFlatDriftDetected(ctx, install, sub, rendered)
			} else {
				err = h.postOrThread(ctx, install, sub, data, rendered, logger)
			}
			if err == nil {
				continue
			}
			sendErrs = append(sendErrs, err)
			logger.Warn("failed to deliver slack lifecycle message",
				zap.String("team_id", link.TeamID),
				zap.String("channel_id", sub.ChannelID),
				zap.Error(err))

			// If Slack reports the workspace is no longer reachable, flip
			// our installation state immediately so subsequent events skip
			// this workspace. We bail on remaining subs for this workspace
			// — they share the same dead token.
			if isSlackUninstallError(err) {
				if mErr := h.markWorkspaceUninstalled(ctx, install.TeamID); mErr != nil {
					logger.Warn("failed to mark slack workspace uninstalled after token failure",
						zap.String("team_id", install.TeamID),
						zap.Error(mErr))
				} else {
					logger.Info("marked slack workspace uninstalled after token failure",
						zap.String("team_id", install.TeamID))
				}
				break
			}
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}
	return nil
}

// postOrThread is the per-subscription dispatcher. It implements the
// (team, channel, workflow) → parent_ts cache:
//
//   - Cache miss: post a parent message (no thread_ts), INSERT the anchor
//     ON CONFLICT DO NOTHING. If we lost the race, re-SELECT and adopt the
//     winner's ts (the orphan parent we posted is acceptable POC degradation
//     and is logged). Then post the child as a threaded reply.
//
//   - Cache hit: post the child as a threaded reply, then best-effort
//     UpdateMessage on the parent with the freshest rollup. UpdateMessage
//     errors are logged but never returned — the child already landed and
//     the parent will catch up on the next event.
func (h *SlackSignalLifecycleHook) postOrThread(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	data lifecycleEventData,
	rendered renderEvent,
	logger *zap.Logger,
) error {
	anchorWFID := anchorWorkflowID(data)
	if anchorWFID == "" {
		// Defensive: without a workflow id we can't thread. Fall back to a
		// flat post.
		flat := slackrender.BuildFlatMessage(rendered.event)
		_, err := h.slackClient.PostMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
			Channel: sub.ChannelID,
			Text:    flat.Text,
			Blocks:  flat.Blocks,
		})
		return err
	}

	anchor, found, err := h.lookupAnchor(ctx, install.TeamID, sub.ChannelID, anchorWFID)
	if err != nil {
		return fmt.Errorf("lookup slack thread anchor: %w", err)
	}

	startedAt := time.Now().UTC()
	parentTS := ""

	if found {
		parentTS = anchor.ParentTS
		// Persisted CreatedAt is the canonical workflow start time. Reading
		// it back keeps elapsed renders consistent across worker replicas.
		startedAt = anchor.CreatedAt
	} else {
		// Cache miss: post the parent first (with no thread_ts).
		parentMsg := slackrender.BuildParentMessage(rendered.event, startedAt)
		parentResp, postErr := h.slackClient.PostMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
			Channel: sub.ChannelID,
			Text:    parentMsg.Text,
			Blocks:  parentMsg.Blocks,
		})
		if postErr != nil {
			return fmt.Errorf("post slack parent message: %w", postErr)
		}
		parentTS = parentResp.TS

		// Persist the anchor. ON CONFLICT DO NOTHING serializes concurrent
		// posts across worker replicas via the unique index on
		// (team_id, channel_id, workflow_id).
		anchorRow := app.SlackThreadAnchor{
			TeamID:       install.TeamID,
			ChannelID:    sub.ChannelID,
			WorkflowID:   anchorWFID,
			ParentTS:     parentTS,
			OrgID:        rendered.event.OrgID,
			WorkflowType: rendered.event.Workflow.Type,
			CreatedAt:    startedAt,
		}
		insertResult := h.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&anchorRow)
		if insertResult.Error != nil {
			logger.Warn("failed to persist slack thread anchor",
				zap.String("team_id", install.TeamID),
				zap.String("channel_id", sub.ChannelID),
				zap.String("workflow_id", anchorWFID),
				zap.Error(insertResult.Error))
			// Continue: child reply will at least land under the parent we
			// just posted, even if we won't be able to consolidate future
			// events under it.
		} else if insertResult.RowsAffected == 0 {
			// We lost the race. Re-SELECT to adopt the winner's ts.
			winner, winnerFound, lookupErr := h.lookupAnchor(ctx, install.TeamID, sub.ChannelID, anchorWFID)
			if lookupErr != nil {
				logger.Warn("failed to re-select slack thread anchor after race",
					zap.Error(lookupErr))
			} else if winnerFound {
				logger.Info("slack thread anchor race lost — adopting winner ts; orphan parent left in channel",
					zap.String("orphan_ts", parentTS),
					zap.String("winner_ts", winner.ParentTS))
				parentTS = winner.ParentTS
				startedAt = winner.CreatedAt
			}
		}
	}

	// Post the child as a threaded reply.
	childMsg := slackrender.BuildChildMessage(rendered.event)
	if _, err := h.slackClient.PostMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel:  sub.ChannelID,
		Text:     childMsg.Text,
		Blocks:   childMsg.Blocks,
		ThreadTS: parentTS,
	}); err != nil {
		return fmt.Errorf("post slack threaded reply: %w", err)
	}

	// Best-effort: edit the parent with the freshest rollup. Failure here
	// is logged but never returned — the child already landed.
	if found {
		rollup := slackrender.BuildParentRollup(rendered.event, startedAt)
		if _, err := h.slackClient.UpdateMessage(ctx, install.BotAccessToken, slackclient.UpdateMessageRequest{
			Channel: sub.ChannelID,
			TS:      parentTS,
			Text:    rollup.Text,
			Blocks:  rollup.Blocks,
		}); err != nil {
			logger.Debug("failed to update slack parent rollup",
				zap.String("team_id", install.TeamID),
				zap.String("channel_id", sub.ChannelID),
				zap.String("parent_ts", parentTS),
				zap.Error(err))
		}
	}

	return nil
}

// postFlatDriftDetected posts a standalone drift notification to a single
// channel subscription. There is no parent anchor, no thread, and no
// rollup edit — each detection lands as its own top-level message that
// links directly to the affected component or sandbox.
func (h *SlackSignalLifecycleHook) postFlatDriftDetected(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildDriftDetectedMessage(rendered.event)
	if _, err := h.slackClient.PostMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack drift-detected message: %w", err)
	}
	return nil
}

// lookupAnchor selects the anchor row for (team, channel, workflow). Returns
// found=false on gorm.ErrRecordNotFound; non-nil error otherwise.
func (h *SlackSignalLifecycleHook) lookupAnchor(ctx context.Context, teamID, channelID, workflowID string) (app.SlackThreadAnchor, bool, error) {
	var anchor app.SlackThreadAnchor
	err := h.db.WithContext(ctx).
		Where(app.SlackThreadAnchor{
			TeamID:     teamID,
			ChannelID:  channelID,
			WorkflowID: workflowID,
		}).
		First(&anchor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app.SlackThreadAnchor{}, false, nil
		}
		return app.SlackThreadAnchor{}, false, err
	}
	return anchor, true, nil
}

// renderEvent bundles the slackrender.Event with the source data so the
// dispatcher can access both without re-translating.
type renderEvent struct {
	event slackrender.Event
}

// buildRenderEvent translates the webhook payload (lifecycleEventData) into
// the slackrender.Event shape the renderer consumes.
func buildRenderEvent(data lifecycleEventData) renderEvent {
	e := slackrender.Event{
		Kind:       data.Kind,
		Transition: data.Transition,
		OrgID:      data.OrgID,
		OrgName:    data.OrgName,
		Workflow: slackrender.WorkflowRef{
			ID:        data.Workflow.ID,
			Type:      data.Workflow.Type,
			OwnerID:   data.Workflow.OwnerID,
			OwnerType: data.Workflow.OwnerType,
			OwnerName: data.Workflow.OwnerName,
		},
	}

	if data.Step != nil {
		e.Step = &slackrender.StepRef{
			ID:            data.Step.ID,
			Name:          data.Step.Name,
			Idx:           data.Step.Idx,
			TargetType:    data.Step.TargetType,
			TargetID:      data.Step.TargetID,
			ComponentID:   data.Step.ComponentID,
			ComponentName: data.Step.ComponentName,
			SandboxID:     data.Step.SandboxID,
			ExecutionType: data.Step.ExecutionType,
		}
	}
	if data.Parent != nil {
		e.Parent = &slackrender.ParentRef{
			WorkflowID: data.Parent.WorkflowID,
			StepID:     data.Parent.StepID,
			Kind:       data.Parent.Kind,
			ActionName: data.Parent.ActionName,
		}
	}
	if data.Outcome != nil {
		e.Outcome = &slackrender.Outcome{
			Status:     data.Outcome.Status,
			Error:      data.Outcome.Error,
			DurationMs: data.Outcome.DurationMs,
		}
	}
	if data.Approval != nil {
		e.Approval = &slackrender.ApprovalRef{
			ID:          data.Approval.ID,
			Type:        data.Approval.Type,
			Plan:        data.Approval.Plan,
			RespondedBy: data.Approval.RespondedBy,
		}
	}
	if data.Links != nil {
		e.Links = &slackrender.ContextLinks{
			Org:        data.Links.Org,
			Install:    data.Links.Install,
			Workflow:   data.Links.Workflow,
			Sandbox:    data.Links.Sandbox,
			Component:  data.Links.Component,
			Approval:   data.Links.Approval,
			RespondAPI: data.Links.RespondAPI,
		}
	}

	return renderEvent{event: e}
}

// anchorWorkflowID resolves the threading anchor: nested action_workflow_run
// sub-workflows consolidate under their launching deploy step's workflow so
// the parent post stays singular for the user-visible run.
func anchorWorkflowID(data lifecycleEventData) string {
	if data.Parent != nil && data.Parent.WorkflowID != "" {
		return data.Parent.WorkflowID
	}
	return data.Workflow.ID
}

// markWorkspaceUninstalled mirrors the Phase 4 events handler's transactional
// uninstall: flip the installation Status, revoke verified org-links, soft-
// delete subscriptions, and hard-delete any thread anchors (their parent ts
// references are dead and re-thread under a stale ts would 404). Used as a
// recovery path when chat.postMessage reports the bot token is dead before
// Slack's lifecycle event reaches us.
func (h *SlackSignalLifecycleHook) markWorkspaceUninstalled(ctx context.Context, teamID string) error {
	if h.db == nil || teamID == "" {
		return nil
	}
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&app.SlackInstallation{}).
			Where(app.SlackInstallation{TeamID: teamID}).
			Updates(map[string]any{
				"status": app.SlackInstallationStatusUninstalled,
			}).Error; err != nil {
			return fmt.Errorf("update installation status: %w", err)
		}
		if err := tx.Model(&app.SlackOrgLink{}).
			Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
			Updates(map[string]any{
				"status": app.SlackOrgLinkStatusRevoked,
			}).Error; err != nil {
			return fmt.Errorf("revoke org links: %w", err)
		}
		if err := tx.Where(app.SlackChannelSubscription{TeamID: teamID}).
			Delete(&app.SlackChannelSubscription{}).Error; err != nil {
			return fmt.Errorf("soft-delete channel subscriptions: %w", err)
		}
		if err := tx.Where(app.SlackThreadAnchor{TeamID: teamID}).
			Delete(&app.SlackThreadAnchor{}).Error; err != nil {
			return fmt.Errorf("delete thread anchors: %w", err)
		}
		return nil
	})
}

// isSlackUninstallError sniffs a Slack client error for the small set of
// strings that indicate the bot token is dead. The handwritten Slack client
// formats errors as `slack chat.postMessage: <slack_err>` so substring match
// is the simplest robust check.
func isSlackUninstallError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "account_inactive") ||
		strings.Contains(msg, "token_revoked") ||
		strings.Contains(msg, "invalid_auth")
}
