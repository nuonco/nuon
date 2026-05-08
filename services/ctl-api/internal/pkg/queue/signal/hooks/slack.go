package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
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
	MW          metrics.Writer      `optional:"true"`
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
	mw          metrics.Writer
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
		mw:          params.MW,
	}
}

// metricNamespace returns the Temporal namespace tag value for metrics emitted
// from inside an activity. Returns "" when called outside an activity context.
func (h *SlackSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

// emitPublishLatency records how long a successful Slack delivery took for
// this phase. Only called when at least one Slack message was actually sent;
// short-circuit / no-subscription paths do not emit so the percentile reflects
// real delivery cost.
func (h *SlackSignalLifecycleHook) emitPublishLatency(ctx context.Context, phasePrefix string, startTS time.Time) {
	if h.mw == nil {
		return
	}
	h.mw.Timing(
		fmt.Sprintf("signal_lifecycle.%s.slack.publish_latency", phasePrefix),
		time.Since(startTS),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

// emitError increments the error counter for this phase. One increment per
// failed delivery / lookup so the count reflects per-attempt failures, not
// per-event.
func (h *SlackSignalLifecycleHook) emitError(ctx context.Context, phasePrefix string) {
	if h.mw == nil {
		return
	}
	h.mw.Incr(
		fmt.Sprintf("signal_lifecycle.%s.slack.errors", phasePrefix),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
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
	// outcome is nil when invoked from BeforePhase, non-nil from AfterPhase.
	// Used as the metric-name prefix so before/after timings are split into
	// separate timeseries (see signal_lifecycle.{before,after}_phase.slack.*).
	phasePrefix := "before_phase"
	if outcome != nil {
		phasePrefix = "after_phase"
	}
	startTS := time.Now()
	delivered := false
	defer func() {
		if delivered {
			h.emitPublishLatency(ctx, phasePrefix, startTS)
		}
	}()

	if event.OrgID == "" || event.WorkflowID == "" {
		return nil
	}

	// Resolve verified org-links and active installations BEFORE the
	// expensive buildEventData enrichment. Both are cheap indexed lookups
	// (org_id / team_id IN); enrichment runs several JOIN queries against
	// install_deploys, install_components, etc. and is wasted work when no
	// Slack workspace is wired up for this org. Most orgs have no Slack
	// integration, so this short-circuit removes the dominant DB cost from
	// the activity's hot path.
	var links []app.SlackOrgLink
	if err := h.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			OrgID:  event.OrgID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		Find(&links).Error; err != nil {
		h.emitError(ctx, phasePrefix)
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
		h.emitError(ctx, phasePrefix)
		return fmt.Errorf("unable to list slack installations for slack lifecycle: %w", err)
	}
	if len(installations) == 0 {
		return nil
	}

	installByTeam := make(map[string]*app.SlackInstallation, len(installations))
	for i := range installations {
		installByTeam[installations[i].TeamID] = &installations[i]
	}

	// Reuse webhook.go's payload builder so the renderer sees exactly the
	// same enriched shape webhook consumers see (plus the same hidden-step
	// suppression rules).
	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	rendered := buildRenderEvent(data)

	// Resolve the entity ids referenced by this event (install / component /
	// action). Drives the per-subscription Match.Matches predicate below.
	// Any of these may be empty for org-only events; in that case only
	// nil-Match subscriptions (org-wide) fire.
	targets := h.eventTargetsFromEvent(ctx, event, data)

	// labelLoader memoises label lookups for this publish() call. Multiple
	// subs in the same channel that hit the same install / component /
	// action only pay the SELECT cost once — events fan out across many
	// publish() invocations for unrelated workflows, so the cache is local
	// to this call rather than a long-lived hook field.
	labelLoader := newLabelLoader(h.db)

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("org_id", event.OrgID),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("anchor_workflow_id", anchorWorkflowID(data)),
		zap.String("event_install_id", targets.InstallID),
		zap.String("event_component_id", targets.ComponentID),
		zap.String("event_action_id", targets.ActionID),
	)

	// Per-channel dedup: at-most-one message per channel per event. Two
	// active subs in the same channel both matching the same event only
	// produce one post. Keyed by channel + queue_signal_id so unrelated
	// events in the same channel still flow.
	seen := make(map[string]struct{})

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
			h.emitError(ctx, phasePrefix)
			continue
		}

		for _, sub := range subs {
			// Resolve labels lazily — the Match predicate may not need
			// them at all (id-only or nil-Match cases). loadEventLabels
			// is idempotent and memoised per-publish.
			if err := labelLoader.load(ctx, &targets); err != nil {
				logger.Warn("failed to load event labels",
					zap.Error(err))
				// Fail open: a label lookup failure shouldn't drop the
				// dispatch. Selector matches will simply miss when the
				// label set is empty.
			}

			if !sub.Match.Matches(targets) {
				continue
			}
			if !interests.Matches(event, outcome, h.db, sub.Interests) {
				continue
			}

			dedupKey := sub.ChannelID + "|" + event.QueueSignalID
			if _, dup := seen[dedupKey]; dup {
				continue
			}
			seen[dedupKey] = struct{}{}

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
				delivered = true
				continue
			}
			sendErrs = append(sendErrs, err)
			h.emitError(ctx, phasePrefix)
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

// eventTargetsFromEvent resolves the entity ids referenced by a lifecycle
// event into the labels.EventTargets shape consumed by SubscriptionMatch.
// Each id is best-effort and may be empty — Match.matches treats an empty
// id as "no entity of this kind on the event" so a component-only event
// never falsely satisfies an installs filter.
//
// Install resolution mirrors the legacy installIDFromEvent path verbatim
// (event.OwnerType, data.Workflow.OwnerType, then step-derived lookups for
// install_deploys / install_sandbox_runs / install_sandboxes). Component
// and action resolution layer alongside without disturbing it.
func (h *SlackSignalLifecycleHook) eventTargetsFromEvent(ctx context.Context, event signal.SignalPhaseEvent, data lifecycleEventData) labels.EventTargets {
	t := labels.EventTargets{}

	// Install id ----------------------------------------------------------
	switch {
	case event.OwnerType == "installs" && event.OwnerID != "":
		t.InstallID = event.OwnerID
	case data.Workflow.OwnerType == "installs" && data.Workflow.OwnerID != "":
		t.InstallID = data.Workflow.OwnerID
	}

	// Component id --------------------------------------------------------
	switch {
	case event.OwnerType == "components" && event.OwnerID != "":
		t.ComponentID = event.OwnerID
	case data.Workflow.OwnerType == "components" && data.Workflow.OwnerID != "":
		t.ComponentID = data.Workflow.OwnerID
	}

	// Action id (action_workflows) ----------------------------------------
	switch {
	case event.OwnerType == "action_workflows" && event.OwnerID != "":
		t.ActionID = event.OwnerID
	case data.Workflow.OwnerType == "action_workflows" && data.Workflow.OwnerID != "":
		t.ActionID = data.Workflow.OwnerID
	}

	// Step-derived enrichment. The enrichment in webhook.go has already
	// surfaced ComponentID and SandboxID on data.Step where applicable; we
	// fan out from those plus the step's TargetType to derive install and
	// action ids.
	if data.Step != nil {
		// Step-surfaced component id wins if not already populated.
		if t.ComponentID == "" && data.Step.ComponentID != "" {
			t.ComponentID = data.Step.ComponentID
		}

		switch data.Step.TargetType {
		case string(app.WorkflowStepTargetTypeInstallDeploy),
			string(app.WorkflowStepTargetTypeInstallDeploys):
			if t.InstallID == "" {
				if id := h.lookupInstallIDFromDeploy(ctx, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		case string(app.WorkflowStepTargetTypeInstallSandboxRun),
			string(app.WorkflowStepTargetTypeInstallSandboxRuns):
			if t.InstallID == "" {
				if id := h.lookupInstallIDFromSandboxRun(ctx, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		case string(app.WorkflowStepTargetTypeInstallActionWorkflowRun),
			string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns):
			if t.ActionID == "" {
				if id := h.lookupActionIDFromInstallActionWorkflowRun(ctx, data.Step.TargetID); id != "" {
					t.ActionID = id
				}
			}
		}

		// Sandbox-derived install. The sandbox is owned by exactly one
		// install.
		if t.InstallID == "" && data.Step.SandboxID != "" {
			if id := h.lookupInstallIDFromSandbox(ctx, data.Step.SandboxID); id != "" {
				t.InstallID = id
			}
		}
	}

	// Parent action: a nested action_workflow_run sub-workflow consolidates
	// under its launching deploy step's workflow. webhook's lookupParent
	// surfaces an action name on data.Parent but not the action id; the
	// id comes from the parent step's target. We don't re-walk that here
	// — Packet C only needs the launching workflow's own action context,
	// which the OwnerType check above already covers.

	return t
}

// lookupActionIDFromInstallActionWorkflowRun resolves the action_workflow_id
// behind an install_action_workflow_runs row by walking through
// install_action_workflows.action_workflow_id. Best-effort: returns "" on
// any DB error or when the row is unlinked (manual triggers may leave
// install_action_workflow_id null).
func (h *SlackSignalLifecycleHook) lookupActionIDFromInstallActionWorkflowRun(ctx context.Context, runID string) string {
	if h.db == nil || runID == "" {
		return ""
	}
	var row struct {
		ActionWorkflowID string
	}
	if err := h.db.WithContext(ctx).
		Table("install_action_workflow_runs").
		Select("install_action_workflows.action_workflow_id AS action_workflow_id").
		Joins("JOIN install_action_workflows ON install_action_workflows.id = install_action_workflow_runs.install_action_workflow_id").
		Where("install_action_workflow_runs.id = ?", runID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.ActionWorkflowID
}

// labelLoader memoises label lookups for one publish() call. Keyed by
// "<kind>:<id>" so install / component / action namespaces never collide.
// The cache lives only for the lifetime of one publish() — events in
// unrelated publish calls don't reuse it.
type labelLoader struct {
	db    *gorm.DB
	cache map[string]labels.Labels
}

func newLabelLoader(db *gorm.DB) *labelLoader {
	return &labelLoader{db: db, cache: make(map[string]labels.Labels)}
}

// load fills in the *Labels fields on t for any (id, labels) combination not
// already populated. Idempotent: calling load multiple times for the same
// targets only triggers a SELECT once per unique entity. Returns the first
// non-nil DB error encountered, but partial population (e.g. install labels
// loaded, components query failed) is still surfaced — the caller is
// expected to fail open and let Matches see whatever it has.
func (l *labelLoader) load(ctx context.Context, t *labels.EventTargets) error {
	if l == nil || l.db == nil || t == nil {
		return nil
	}
	var firstErr error
	if t.InstallID != "" && t.InstallLabels == nil {
		lbls, err := l.fetch(ctx, "installs", t.InstallID)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		t.InstallLabels = lbls
	}
	if t.ComponentID != "" && t.ComponentLabels == nil {
		lbls, err := l.fetch(ctx, "components", t.ComponentID)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		t.ComponentLabels = lbls
	}
	if t.ActionID != "" && t.ActionLabels == nil {
		lbls, err := l.fetch(ctx, "action_workflows", t.ActionID)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		t.ActionLabels = lbls
	}
	return firstErr
}

// fetch reads `labels` from the table and memoises the result. A cache hit
// returns the cached value (which may be an empty Labels{} for a row with
// no labels). A cache miss consults the DB and stores a defensive
// non-nil Labels{} on success — distinguishes "we looked, none set" from
// "haven't loaded yet".
func (l *labelLoader) fetch(ctx context.Context, table, id string) (labels.Labels, error) {
	key := table + ":" + id
	if cached, ok := l.cache[key]; ok {
		return cached, nil
	}
	var row struct {
		Labels labels.Labels
	}
	if err := l.db.WithContext(ctx).
		Table(table).
		Select("labels").
		Where("id = ?", id).
		Scan(&row).Error; err != nil {
		// Cache the miss too so a transient error doesn't multiply
		// queries during the same publish.
		l.cache[key] = labels.Labels{}
		return labels.Labels{}, err
	}
	if row.Labels == nil {
		row.Labels = labels.Labels{}
	}
	l.cache[key] = row.Labels
	return row.Labels, nil
}

func (h *SlackSignalLifecycleHook) lookupInstallIDFromDeploy(ctx context.Context, deployID string) string {
	if h.db == nil || deployID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := h.db.WithContext(ctx).
		Table("install_deploys").
		Select("install_components.install_id AS install_id").
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_deploys.id = ?", deployID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

func (h *SlackSignalLifecycleHook) lookupInstallIDFromSandboxRun(ctx context.Context, sandboxRunID string) string {
	if h.db == nil || sandboxRunID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := h.db.WithContext(ctx).
		Table("install_sandbox_runs").
		Select("install_id").
		Where("id = ?", sandboxRunID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

func (h *SlackSignalLifecycleHook) lookupInstallIDFromSandbox(ctx context.Context, sandboxID string) string {
	if h.db == nil || sandboxID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := h.db.WithContext(ctx).
		Table("install_sandboxes").
		Select("install_id").
		Where("id = ?", sandboxID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
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
