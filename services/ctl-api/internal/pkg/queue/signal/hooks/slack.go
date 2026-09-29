package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"
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

type SlackParams struct {
	fx.In

	Cfg           *internal.Config     `optional:"true"`
	L             *zap.Logger          `optional:"true"`
	DB            *gorm.DB             `name:"psql" optional:"true"`
	SlackClient   *slackclient.Client  `optional:"true"`
	MW            metrics.Writer       `optional:"true"`
	MeterProvider metric.MeterProvider `optional:"true"`
}

type SlackSignalLifecycleHook struct {
	l               *zap.Logger
	db              *gorm.DB
	slackClient     *slackclient.Client
	appURL          string
	enricher        *WebhookSignalLifecycleHook
	mw              metrics.Writer
	deliveryMetrics *deliveryMetrics
}

var _ signal.SignalLifecycleHook = (*SlackSignalLifecycleHook)(nil)

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

	enricher := &WebhookSignalLifecycleHook{
		l:            logger,
		db:           params.DB,
		appURL:       appURL,
		publicAPIURL: publicAPIURL,
	}

	return &SlackSignalLifecycleHook{
		l:               logger,
		db:              params.DB,
		slackClient:     params.SlackClient,
		appURL:          appURL,
		enricher:        enricher,
		mw:              params.MW,
		deliveryMetrics: newDeliveryMetrics(params.MeterProvider),
	}
}

func (h *SlackSignalLifecycleHook) postMessage(ctx context.Context, botToken string, req slackclient.PostMessageRequest) (resp *slackclient.PostMessageResponse, retErr error) {
	started := time.Now()
	defer func() {
		h.deliveryMetrics.record(ctx, deliveryChannelSlack, deliveryOperationPost, started, retErr)
	}()
	return h.slackClient.PostMessage(ctx, botToken, req)
}

func (h *SlackSignalLifecycleHook) updateMessage(ctx context.Context, botToken string, req slackclient.UpdateMessageRequest) (resp *slackclient.UpdateMessageResponse, retErr error) {
	started := time.Now()
	defer func() {
		h.deliveryMetrics.record(ctx, deliveryChannelSlack, deliveryOperationUpdate, started, retErr)
	}()
	return h.slackClient.UpdateMessage(ctx, botToken, req)
}

func (h *SlackSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

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

func (h *SlackSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.slackClient == nil || h.db == nil {
		return false
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow,
		signalTypeExecuteWorkflowStep,
		signalTypeWorkflowStepApprovalRequest,
		signalTypeWorkflowStepApprovalResponse,
		signalTypeDriftDetected,
		signalTypeWorkflowStepAwaitingRetry,
		signalTypeStackRun,
		signalTypeRoleChange,
		signalTypeInputsUpdated,
		signalTypeAppConfigSynced,
		signalTypeUpdateAppConfig,
		signalTypeComponentUnhealthy,
		signalTypeComponentRecovered,
		signalTypeInstallDegraded,
		signalTypeRunnerUnhealthy,
		signalTypeSyncInstalls,
		signalTypeInstallConfigSync,
		signalTypeLabelAdded:
		return true
	default:
		return false
	}
}

func (h *SlackSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	if suppressesStartedEvent(event.SignalType) {
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

	suppress, err := resolveFlowCompletionOutcome(ctx, workflowStatusFromDB(h.db), event, &outcome)
	if err != nil {
		return fmt.Errorf("unable to resolve workflow outcome before lifecycle completion: %w", err)
	}
	if suppress {
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

func (h *SlackSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
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

	if event.OrgID == "" {
		return nil
	}
	if event.WorkflowID == "" && !isNotificationOnlySignalType(event.SignalType) {
		return nil
	}

	var links []app.SlackOrgLink
	if err := retryDBRead(ctx, func() error {
		return h.db.WithContext(ctx).
			Where(app.SlackOrgLink{
				OrgID:  event.OrgID,
				Status: app.SlackOrgLinkStatusVerified,
			}).
			Find(&links).Error
	}); err != nil {
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
	if err := retryDBRead(ctx, func() error {
		return h.db.WithContext(ctx).
			Where("team_id IN ? AND status = ?", teamIDs, app.SlackInstallationStatusActive).
			Find(&installations).Error
	}); err != nil {
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

	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	if data.Kind == kindWorkflowStep && data.Transition == transitionSucceeded && event.StepID != "" {
		if approval, ok := h.enricher.lookupStepApproval(ctx, event.StepID); ok {
			if resp, ok := h.enricher.lookupApprovalResponse(ctx, approval.ID); ok {
				if mapApprovalResponseTransition(resp.Type) == transitionRejected {
					return nil
				}
			}
		}
	}

	if event.SignalType == signalTypeRoleChange {
		h.enrichRoleChangeWithActionTriggers(ctx, event, &data)
	}

	rendered := buildRenderEvent(data)

	targets := h.eventTargetsFromEvent(ctx, event, data)

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

	seen := make(map[string]struct{})

	var sendErrs []error
	for _, link := range links {
		install, ok := installByTeam[link.TeamID]
		if !ok {
			continue
		}

		var subs []app.SlackChannelSubscription
		if err := retryDBRead(ctx, func() error {
			return h.db.WithContext(ctx).
				Where(app.SlackChannelSubscription{
					OrgLinkID: link.ID,
					OrgID:     event.OrgID,
				}).
				Find(&subs).Error
		}); err != nil {
			logger.Warn("failed to list channel subscriptions",
				zap.String("team_id", link.TeamID), zap.Error(err))
			sendErrs = append(sendErrs, err)
			h.emitError(ctx, phasePrefix)
			continue
		}

		for _, sub := range subs {
			if err := labelLoader.load(ctx, &targets); err != nil {
				logger.Warn("failed to load event labels",
					zap.Error(err))
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

			var err error
			if isNotificationOnlySignalType(event.SignalType) {
				err = h.postFlatNotification(ctx, install, sub, rendered, event.SignalType)
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
		flat := slackrender.BuildFlatMessage(rendered.event)
		_, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
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
	if !rendered.event.Workflow.CreatedAt.IsZero() {
		startedAt = rendered.event.Workflow.CreatedAt.UTC()
	}
	parentTS := ""

	if found {
		parentTS = anchor.ParentTS
		startedAt = anchor.CreatedAt
	} else {
		parentMsg := slackrender.BuildParentMessage(rendered.event, startedAt)
		parentResp, postErr := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
			Channel: sub.ChannelID,
			Text:    parentMsg.Text,
			Blocks:  parentMsg.Blocks,
		})
		if postErr != nil {
			return fmt.Errorf("post slack parent message: %w", postErr)
		}
		parentTS = parentResp.TS

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
		} else if insertResult.RowsAffected == 0 {
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

	if rendered.event.Kind != slackrender.KindWorkflow {
		childMsg := slackrender.BuildChildMessage(rendered.event)
		if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
			Channel:  sub.ChannelID,
			Text:     childMsg.Text,
			Blocks:   childMsg.Blocks,
			ThreadTS: parentTS,
		}); err != nil {
			return fmt.Errorf("post slack threaded reply: %w", err)
		}
	}

	if found {
		rollup := slackrender.BuildParentRollup(rendered.event, startedAt)
		if _, err := h.updateMessage(ctx, install.BotAccessToken, slackclient.UpdateMessageRequest{
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

func (h *SlackSignalLifecycleHook) postFlatDriftDetected(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildDriftDetectedMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack drift-detected message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatRoleChange(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildRoleChangeMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack role-change message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatAppConfigSynced(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildAppConfigSyncedMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack app-config-synced message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatUpdateAppConfig(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildUpdateAppConfigMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack update-app-config message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatComponentHealth(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
	signalType signal.SignalType,
) error {
	msg := slackrender.BuildComponentHealthMessage(
		rendered.event,
		signalType == signalTypeComponentRecovered,
		signalType == signalTypeInstallDegraded,
	)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack component-health message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatRunnerUnhealthy(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
) error {
	msg := slackrender.BuildRunnerUnhealthyMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack runner-unhealthy message: %w", err)
	}
	return nil
}

func (h *SlackSignalLifecycleHook) postFlatNotification(
	ctx context.Context,
	install *app.SlackInstallation,
	sub app.SlackChannelSubscription,
	rendered renderEvent,
	signalType signal.SignalType,
) error {
	if signalType == signalTypeDriftDetected {
		return h.postFlatDriftDetected(ctx, install, sub, rendered)
	}
	if signalType == signalTypeRoleChange {
		return h.postFlatRoleChange(ctx, install, sub, rendered)
	}
	if signalType == signalTypeAppConfigSynced {
		return h.postFlatAppConfigSynced(ctx, install, sub, rendered)
	}
	if signalType == signalTypeUpdateAppConfig {
		return h.postFlatUpdateAppConfig(ctx, install, sub, rendered)
	}
	if signalType == signalTypeRunnerUnhealthy {
		return h.postFlatRunnerUnhealthy(ctx, install, sub, rendered)
	}
	if isComponentHealthSignalType(signalType) {
		return h.postFlatComponentHealth(ctx, install, sub, rendered, signalType)
	}
	msg := slackrender.BuildFlatMessage(rendered.event)
	if _, err := h.postMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: sub.ChannelID,
		Text:    msg.Text,
		Blocks:  msg.Blocks,
	}); err != nil {
		return fmt.Errorf("post slack %s message: %w", signalType, err)
	}
	return nil
}

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

type renderEvent struct {
	event slackrender.Event
}

func buildRenderEvent(data lifecycleEventData) renderEvent {
	e := slackrender.Event{
		Kind:       data.Kind,
		Transition: data.Transition,
		OrgID:      data.OrgID,
		OrgName:    data.OrgName,
		Workflow: slackrender.WorkflowRef{
			ID:             data.Workflow.ID,
			Type:           data.Workflow.Type,
			OwnerID:        data.Workflow.OwnerID,
			OwnerType:      data.Workflow.OwnerType,
			OwnerName:      data.Workflow.OwnerName,
			CreatedByEmail: data.Workflow.CreatedByEmail,
			CreatedAt:      data.Workflow.CreatedAt,
			RunbookName:    data.Workflow.RunbookName,
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

	e.Metadata = data.Metadata

	return renderEvent{event: e}
}

func anchorWorkflowID(data lifecycleEventData) string {
	if data.Parent != nil && data.Parent.WorkflowID != "" {
		return data.Parent.WorkflowID
	}
	return data.Workflow.ID
}

func (h *SlackSignalLifecycleHook) eventTargetsFromEvent(ctx context.Context, event signal.SignalPhaseEvent, data lifecycleEventData) labels.EventTargets {
	t := labels.EventTargets{}

	switch {
	case event.OwnerType == "installs" && event.OwnerID != "":
		t.InstallID = event.OwnerID
	case data.Workflow.OwnerType == "installs" && data.Workflow.OwnerID != "":
		t.InstallID = data.Workflow.OwnerID
	}

	switch {
	case event.OwnerType == "components" && event.OwnerID != "":
		t.ComponentID = event.OwnerID
	case data.Workflow.OwnerType == "components" && data.Workflow.OwnerID != "":
		t.ComponentID = data.Workflow.OwnerID
	}

	switch {
	case event.OwnerType == "action_workflows" && event.OwnerID != "":
		t.ActionID = event.OwnerID
	case data.Workflow.OwnerType == "action_workflows" && data.Workflow.OwnerID != "":
		t.ActionID = data.Workflow.OwnerID
	}

	if data.Step != nil {
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
		case string(app.WorkflowStepTargetTypeInstallStackVersions):
			if t.InstallID == "" {
				if id := h.lookupInstallIDFromStackVersion(ctx, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		}

		if t.InstallID == "" && data.Step.SandboxID != "" {
			if id := h.lookupInstallIDFromSandbox(ctx, data.Step.SandboxID); id != "" {
				t.InstallID = id
			}
		}
	}

	return t
}

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

type labelLoader struct {
	db    *gorm.DB
	cache map[string]labels.Labels
}

func newLabelLoader(db *gorm.DB) *labelLoader {
	return &labelLoader{db: db, cache: make(map[string]labels.Labels)}
}

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
		l.cache[key] = labels.Labels{}
		return labels.Labels{}, err
	}
	if row.Labels == nil {
		row.Labels = labels.Labels{}
	}
	l.cache[key] = row.Labels
	return row.Labels, nil
}

func (h *SlackSignalLifecycleHook) lookupInstallIDFromStackVersion(ctx context.Context, stackVersionID string) string {
	if h.db == nil || stackVersionID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := h.db.WithContext(ctx).
		Table("install_stack_versions").
		Select("install_id").
		Where("id = ?", stackVersionID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
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

func (h *SlackSignalLifecycleHook) enrichRoleChangeWithActionTriggers(ctx context.Context, event signal.SignalPhaseEvent, data *lifecycleEventData) {
	if h.db == nil || event.OwnerID == "" {
		return
	}

	changeType, _ := data.Metadata["change_type"].(string)
	triggerType := "role-enabled"
	if changeType == "disabled" {
		triggerType = "role-disabled"
	}

	var appID string
	if err := h.db.WithContext(ctx).
		Table("installs").
		Select("app_id").
		Where("id = ?", event.OwnerID).
		Scan(&appID).Error; err != nil || appID == "" {
		return
	}

	var names []string
	if err := h.db.WithContext(ctx).
		Table("action_workflows").
		Select("action_workflows.name").
		Joins("JOIN action_workflow_configs_latest_view_v1 awc ON awc.action_workflow_id = action_workflows.id").
		Joins("JOIN action_workflow_trigger_configs awtc ON awtc.action_workflow_config_id = awc.id").
		Where("action_workflows.app_id = ? AND awtc.type = ? AND awtc.deleted_at = 0", appID, triggerType).
		Scan(&names).Error; err != nil || len(names) == 0 {
		return
	}

	if data.Metadata == nil {
		data.Metadata = map[string]any{}
	}
	data.Metadata["action_trigger_names"] = strings.Join(names, ", ")
}

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

func isSlackUninstallError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "account_inactive") ||
		strings.Contains(msg, "token_revoked") ||
		strings.Contains(msg, "invalid_auth")
}
