package permissions

const (
	OrgUpdate           PermissionID = "org.update"
	OrgDelete           PermissionID = "org.delete"
	OrgTelemetryUpdate  PermissionID = "org.telemetry.update"
	OrgMembersRead      PermissionID = "org.members.read"
	OrgMemberAdd        PermissionID = "org.member.add"
	OrgMemberRemove     PermissionID = "org.member.remove"
	OrgMemberRoleUpdate PermissionID = "org.member.role.update"
	OrgInvitesRead      PermissionID = "org.invites.read"
	OrgInviteCreate     PermissionID = "org.invite.create"
	OrgInviteRevoke     PermissionID = "org.invite.revoke"
	OrgStatsRead        PermissionID = "org.stats.read"
	OrgFeaturesRead     PermissionID = "org.features.read"
	OrgFeaturesUpdate   PermissionID = "org.features.update"
	OrgWebhooksRead     PermissionID = "org.webhooks.read"
	OrgWebhookCreate    PermissionID = "org.webhook.create"
	OrgWebhookUpdate    PermissionID = "org.webhook.update"
	OrgWebhookDelete    PermissionID = "org.webhook.delete"
	OnboardingManage    PermissionID = "onboarding.manage"
)

func OrgDefinitions() []PermissionDefinition {
	org := []ResourceKind{KindOrg}
	return []PermissionDefinition{
		{ID: OrgWrite, Title: "Org write", Description: "Organization-wide writes, including deletion, automation, credential issuance and access administration. Not an Editor preset.", Scopes: org, Members: []PermissionID{
			OrgUpdate, OrgDelete, OrgTelemetryUpdate, OrgMemberAdd, OrgMemberRemove, OrgMemberRoleUpdate,
			OrgInviteCreate, OrgInviteRevoke, OrgFeaturesUpdate, OrgWebhookCreate, OrgWebhookUpdate, OrgWebhookDelete,
			AppCreate, AppDelete, AppAutomationManage, AppDefinitionWrite, AppBundleCreate, AppBundleDownload,
			InstallCreate, InstallWrite, ActionCreate, ActionUpdate, ActionDelete, ActionConfigCreate, ActionRunLogsDelete,
			RunbookCreate, RunbookUpdate, RunbookDelete, RunbookConfigCreate, OnboardingManage,
			CloudConnectionCreate, CloudConnectionDelete, CloudConnectionVerify, ImageRepositoryCreate, ImageRepositoryDelete,
			VCSConnectionCreate, VCSConnectionDelete, VCSWebhookSubscribe, SlackConnect, SlackDisconnect, SlackSubscriptionWrite,
			TriggerCreate, TriggerDelete, TriggerCredentials, TriggerUpdate, TriggerReplay, TriggerRetry,
			ServiceAccountCreate, ServiceAccountUpdate, ServiceAccountRoleUpdate, ServiceAccountDelete,
			TokenCreate, TokenDelete, TrustPolicyCreate, TrustPolicyUpdate, TrustPolicyDelete,
		}},
		{ID: OnboardingManage, Title: "Complete onboarding", Description: "Create and configure apps and installs through the account's onboarding session in this org, including queued setup work.", Scopes: org, Targets: org},
		{ID: OrgUpdate, Title: "Update org", Description: "Update organization metadata and settings.", Scopes: org, Targets: org},
		{ID: OrgDelete, Title: "Delete org", Description: "Delete an organization and its existing descendant cleanup, subject to existing constraints.", Scopes: org, Targets: org},
		{ID: OrgTelemetryUpdate, Title: "Update org telemetry", Description: "Change organization telemetry and relay settings; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: OrgMembersRead, Title: "Read org members", Description: "Read organization accounts and member role metadata.", Scopes: org, Targets: org},
		{ID: OrgMemberAdd, Title: "Add org membership", Description: "Use the existing self-membership endpoint, including its org-admin assignment.", Scopes: org, Targets: org},
		{ID: OrgMemberRemove, Title: "Remove org members", Description: "Remove a member's organization bindings and clean up their organization credentials.", Scopes: org, Targets: org},
		{ID: OrgMemberRoleUpdate, Title: "Change member roles", Description: "Assign existing team roles, preserving self-demotion and last-admin constraints; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: OrgInvitesRead, Title: "Read org invites", Description: "Read organization invitations and their roles.", Scopes: org, Targets: org},
		{ID: OrgInviteCreate, Title: "Invite org members", Description: "Create or resend invitations for existing assignable team roles.", Scopes: org, Targets: org},
		{ID: OrgInviteRevoke, Title: "Revoke org invites", Description: "Revoke pending invitations; formerly requires org admin.", Scopes: org, Targets: org},
		{ID: OrgStatsRead, Title: "Read org statistics", Description: "Read organization-wide app and install aggregates.", Scopes: org, Targets: org},
		{ID: OrgFeaturesRead, Title: "Read org features", Description: "Read available and enabled organization features.", Scopes: org, Targets: org},
		{ID: OrgFeaturesUpdate, Title: "Update org features", Description: "Change user-manageable organization features subject to existing feature eligibility.", Scopes: org, Targets: org},
		{ID: OrgWebhooksRead, Title: "Read org webhooks", Description: "Read organization webhook configuration and destinations.", Scopes: org, Targets: org},
		{ID: OrgWebhookCreate, Title: "Create org webhooks", Description: "Create webhook delivery destinations for organization events.", Scopes: org, Targets: org},
		{ID: OrgWebhookUpdate, Title: "Update org webhooks", Description: "Change organization webhook destinations and event subscriptions.", Scopes: org, Targets: org},
		{ID: OrgWebhookDelete, Title: "Delete org webhooks", Description: "Delete organization webhook delivery configuration.", Scopes: org, Targets: org},
	}
}
