package permissions

const (
	CloudConnectionRead    PermissionID = "cloud_connection.read"
	CloudConnectionCreate  PermissionID = "cloud_connection.create"
	CloudConnectionDelete  PermissionID = "cloud_connection.delete"
	CloudConnectionVerify  PermissionID = "cloud_connection.verify"
	ImageRepositoryRead    PermissionID = "image_repository.read"
	ImageRepositoryCreate  PermissionID = "image_repository.create"
	ImageRepositoryDelete  PermissionID = "image_repository.delete"
	VCSConnectionRead      PermissionID = "vcs_connection.read"
	VCSConnectionCreate    PermissionID = "vcs_connection.create"
	VCSConnectionDelete    PermissionID = "vcs_connection.delete"
	VCSWebhookSubscribe    PermissionID = "vcs_webhook.subscribe"
	SlackRead              PermissionID = "slack.read"
	SlackConnect           PermissionID = "slack.connect"
	SlackDisconnect        PermissionID = "slack.disconnect"
	SlackSubscriptionWrite PermissionID = "slack.subscription.write"
)

func IntegrationDefinitions() []PermissionDefinition {
	org := []ResourceKind{KindOrg}
	return []PermissionDefinition{
		{ID: CloudConnectionRead, Title: "Read cloud connections", Description: "Read organization cloud connections, usage counts and generated setup configuration.", Scopes: org, Targets: org},
		{ID: CloudConnectionCreate, Title: "Create cloud connections", Description: "Create a cloud connection and request its existing provider verification.", Scopes: org, Targets: org},
		{ID: CloudConnectionDelete, Title: "Delete cloud connections", Description: "Delete unreferenced cloud connections subject to existing usage constraints.", Scopes: org, Targets: org},
		{ID: CloudConnectionVerify, Title: "Verify cloud connections", Description: "Request provider verification for an organization cloud connection.", Scopes: org, Targets: org},
		{ID: ImageRepositoryRead, Title: "Read image repositories", Description: "Read organization image repository assignments and their embedded cloud connections.", Scopes: org, Targets: org},
		{ID: ImageRepositoryCreate, Title: "Assign image repositories", Description: "Assign an image repository to an organization cloud connection for subsequent builds.", Scopes: org, Targets: org},
		{ID: ImageRepositoryDelete, Title: "Remove image repositories", Description: "Remove repository assignments without revoking existing cloud permissions or pinned configurations.", Scopes: org, Targets: org},
		{ID: VCSConnectionRead, Title: "Read VCS connections", Description: "Read organization connections, provider repositories and branches, status and shared webhook subscription metadata.", Scopes: org, Targets: org},
		{ID: VCSConnectionCreate, Title: "Connect VCS", Description: "Connect a provider installation to the selected org after the existing provider ownership check.", Scopes: org, Targets: org},
		{ID: VCSConnectionDelete, Title: "Delete VCS connections", Description: "Delete organization connections and optionally uninstall the provider app, preserving existing shared-installation effects.", Scopes: org, Targets: org},
		{ID: VCSWebhookSubscribe, Title: "Subscribe VCS webhooks", Description: "Enable provider event delivery for a connection and its configured automation.", Scopes: org, Targets: org},
		{ID: SlackRead, Title: "Read Slack integration", Description: "Read organization installations, provider channels, links and notification subscriptions.", Scopes: org, Targets: org},
		{ID: SlackConnect, Title: "Connect Slack", Description: "Issue the signed OAuth state used to connect a Slack installation to the organization.", Scopes: org, Targets: org},
		{ID: SlackDisconnect, Title: "Disconnect Slack", Description: "Remove an organization's Slack installation link and its existing subscription cleanup.", Scopes: org, Targets: org},
		{ID: SlackSubscriptionWrite, Title: "Manage Slack subscriptions", Description: "Create, update and delete organization event delivery subscriptions, including their resource filters.", Scopes: org, Targets: org},
	}
}
