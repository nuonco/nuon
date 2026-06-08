package migrations

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"

func (m *Migrations) All() []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "01-create-internal-accounts",
			Fn:   m.migration01InternalAccounts,
		},
		{
			Name: "002-drop-old-actions-run-index",
			Fn:   m.Migration002DropOldActionsRunIndex,
		},
		{
			Name: "086-runner-group-settings-backfill-groups",
			Fn:   m.Migration086RunnerGroupSettingsBackfillGroups,
		},
		{
			Name: "04",
			SQL:  `ALTER TABLE action_workflow_step_configs ALTER COLUMN command DROP NOT NULL;`,
		},
		{
			Name: "087-install-workflows-backfill-ownership",
			Fn:   m.Migration087InstallWorkflowsBackfillOwnership,
		},
		{
			Name: "088-accounts-email-not-empty",
			Fn:   m.Migration088AccountsEmailsNotEmpty,
		},
		{
			Name: "089-app-branches-cleanup",
			Fn:   m.Migration089AppBracnhesCleanup,
		},
		{
			Name: "090-workflow-step-null-install-id",
			Fn:   m.Migration09NullWorkflowInstallID,
		},
		{
			Name: "091-delete-orphaned-action-triggers",
			Fn:   m.Migration091DeleteOrphanedActionTriggers,
		},
		{
			Name: "092-backfill-org-id",
			Fn:   m.Migration092BackfillOrgID,
		},
		{
			Name: "093-add-adhoc-actions",
			Fn:   m.Migration093AddAdhocActions,
		},
		{
			Name: "094-vcs-commit-polymorphic-ownership",
			Fn:   m.Migration094VCSCommitPolymorphicOwnership,
		},
		{
			Name: "095-backfill-org-support-role",
			Fn:   m.Migration095BackfillOrgSupportRole,
		},
		{
			Name: "096-backfill-install-sandbox-mode",
			Fn:   m.Migration096BackfillInstallSandboxMode,
		},
		{
			Name: "097-backfill-runner-group-owner-name",
			Fn:   m.Migration097BackfillRunnerGroupOwnerName,
		},
		{
			Name: "098-backfill-queue-signal-enqueue-finished-at",
			Fn:   m.Migration098BackfillQueueSignalEnqueueFinishedAt,
		},
		{
			Name: "099-app-config-backfill-status-v2",
			Fn:   m.Migration099AppConfigBackfillStatusV2,
		},
		{
			Name: "100-fix-approval-option-default",
			Fn:   m.Migration100FixApprovalOptionDefault,
		},
		{
			Name: "101-slack-channel-subs-creator-check",
			Fn:   m.Migration101SlackChannelSubsCreatorCheck,
		},
		{
			Name: "102-webhooks-match",
			Fn:   m.Migration102WebhooksMatch,
		},
		{
			Name: "103-queue-signals-inflight-index",
			Fn:   m.Migration103QueueSignalsInflightIndex,
		},
		{
			Name: "104-fix-stuck-generate-workflow-steps-signals",
			Fn:   m.Migration104FixStuckGenerateWorkflowStepsSignals,
		},
		{
			Name: "105-fix-skip-noops-default",
			Fn:   m.Migration105FixSkipNoopsDefault,
		},
		{
			Name: "106-backfill-queue-signal-expires-at",
			Fn:   m.Migration106BackfillQueueSignalExpiresAt,
		},
		{
			Name: "107-backfill-emitter-signal-expires-in",
			Fn:   m.Migration107BackfillEmitterSignalExpiresIn,
		},
		{
			Name: "108-install-workflows-name-hook-managed",
			Fn:   m.Migration108InstallWorkflowsNameHookManaged,
		},
		{
			Name: "109-backfill-runbook-step-deploy-dependents",
			Fn:   m.Migration109BackfillRunbookStepDeployDependents,
		},
		{
			Name: "110-canonicalize-runbook-step-deploy-type",
			Fn:   m.Migration110CanonicalizeRunbookStepDeployType,
		},
		{
			Name: "111-backfill-lifecycle-phase",
			Fn:   m.Migration111BackfillLifecyclePhase,
		},
		{
			Name: "112-runner-job-available-notify-trigger",
			Fn:   m.Migration112RunnerJobAvailableNotifyTrigger,
		},
	}
}
