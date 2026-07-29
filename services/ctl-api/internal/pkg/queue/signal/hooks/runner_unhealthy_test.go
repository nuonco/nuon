package hooks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

func runnerUnhealthyEvent() signal.SignalPhaseEvent {
	installID := "ins_1"
	return signal.SignalPhaseEvent{
		QueueSignalID: "sig_1",
		SignalType:    signalTypeRunnerUnhealthy,
		Phase:         signal.SignalPhaseExecute,
		OrgID:         "org_1",
		OrgName:       "Acme",
		InstallID:     &installID,
		OwnerID:       installID,
		OwnerType:     "installs",
		OwnerName:     "Production",
		Metadata: map[string]any{
			"runner_id":         "run_1",
			"runner_name":       "Default runner",
			"runner_group_id":   "rug_1",
			"runner_group_type": "install",
			"from_status":       "active",
			"to_status":         "offline",
			"reason":            "no active install process",
		},
	}
}

func TestRunnerUnhealthyEventData(t *testing.T) {
	hook := &WebhookSignalLifecycleHook{appURL: "https://app.nuon.co"}
	event := runnerUnhealthyEvent()
	outcome := &signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess}

	data, ok := hook.buildEventData(context.Background(), event, outcome)
	require.True(t, ok)
	assert.Equal(t, kindRunnerUnhealthy, data.Kind)
	assert.Equal(t, transitionUnhealthy, data.Transition)
	assert.Equal(t, "ins_1", data.Workflow.OwnerID)
	assert.Equal(t, "installs", data.Workflow.OwnerType)
	require.NotNil(t, data.Outcome)
	assert.Equal(t, statusFailed, data.Outcome.Status)
	assert.Equal(t, "no active install process", data.Outcome.Error)
	assert.Equal(t, "run_1", data.Metadata["runner_id"])
	assert.Equal(t, "active", data.Metadata["from_status"])
	assert.Equal(t, "offline", data.Metadata["to_status"])
	assert.Equal(t, "org_1/runner_unhealthy/run_1/unhealthy", buildSubject(event, data))
	require.NotNil(t, data.Links)
	assert.Equal(t, "https://app.nuon.co/org_1/installs/ins_1", data.Links.Install)

	targets := EventTargetsFromEvent(context.Background(), nil, event, data)
	assert.Equal(t, "ins_1", targets.InstallID)
	assert.True(t, (&labels.SubscriptionMatch{Installs: &labels.TargetMatch{IDs: []string{"ins_1"}}}).Matches(targets))
	assert.False(t, (&labels.SubscriptionMatch{Installs: &labels.TargetMatch{IDs: []string{"ins_2"}}}).Matches(targets))
}

func TestRunnerUnhealthyOnlyEmitsSuccessfulAfterPhase(t *testing.T) {
	webhookHook := &WebhookSignalLifecycleHook{db: &gorm.DB{}}
	slackHook := &SlackSignalLifecycleHook{db: &gorm.DB{}, slackClient: &slackclient.Client{}}
	event := runnerUnhealthyEvent()

	assert.True(t, webhookHook.Supports(event))
	assert.True(t, slackHook.Supports(event))

	_, ok := webhookHook.buildEventData(context.Background(), event, nil)
	assert.False(t, ok)
	_, ok = webhookHook.buildEventData(context.Background(), event, &signal.SignalPhaseOutcome{Status: signal.SignalStatusError})
	assert.False(t, ok)
	assert.True(t, suppressesStartedEvent(signalTypeRunnerUnhealthy))
	assert.True(t, isNotificationOnlySignalType(signalTypeRunnerUnhealthy))
}

func TestOrgRunnerUnhealthyOnlyMatchesOrgWideSubscriptions(t *testing.T) {
	event := runnerUnhealthyEvent()
	event.InstallID = nil
	event.OwnerID = event.OrgID
	event.OwnerType = "orgs"
	event.OwnerName = event.OrgName

	data, ok := (&WebhookSignalLifecycleHook{}).buildEventData(
		context.Background(),
		event,
		&signal.SignalPhaseOutcome{Status: signal.SignalStatusSuccess},
	)
	require.True(t, ok)
	targets := EventTargetsFromEvent(context.Background(), nil, event, data)
	assert.Empty(t, targets.InstallID)
	assert.True(t, (*labels.SubscriptionMatch)(nil).Matches(targets))
	assert.False(t, (&labels.SubscriptionMatch{Installs: &labels.TargetMatch{}}).Matches(targets))
}
