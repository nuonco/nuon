package permissions

const (
	TriggerRead        PermissionID = "trigger.read"
	TriggerCreate      PermissionID = "trigger.create"
	TriggerDelete      PermissionID = "trigger.delete"
	TriggerCredentials PermissionID = "trigger.credentials.manage"
	TriggerUpdate      PermissionID = "trigger.update"
	TriggerReplay      PermissionID = "trigger.replay"
	TriggerRetry       PermissionID = "trigger.dispatch.retry"
)

func TriggerDefinitions() []PermissionDefinition {
	org := []ResourceKind{KindOrg}
	return []PermissionDefinition{
		{ID: TriggerRead, Title: "Read triggers and events", Description: "Read org triggers, active rules, full event payloads and dispatch metadata across apps and installs.", Scopes: org, Targets: org},
		{ID: TriggerCreate, Title: "Create triggers", Description: "Create org event ingress and its initial credentials.", Scopes: org, Targets: org},
		{ID: TriggerDelete, Title: "Delete triggers", Description: "Remove org event ingress and revoke credentials, including force deletion of referenced triggers under existing waiter constraints.", Scopes: org, Targets: org},
		{ID: TriggerCredentials, Title: "Manage trigger credentials", Description: "Reveal or rotate org event-ingress URLs and signing secrets, or revoke secrets; these credentials can dispatch work through configured rules.", Scopes: org, Targets: org},
		{ID: TriggerUpdate, Title: "Enable or disable triggers", Description: "Enable or disable org ingress and its configured app and install automation.", Scopes: org, Targets: org},
		{ID: TriggerReplay, Title: "Replay org events", Description: "Route a previously accepted event through the org's current rules and waiters, potentially dispatching work across apps and installs.", Scopes: org, Targets: org},
		{ID: TriggerRetry, Title: "Retry event dispatches", Description: "Retry failed dispatches using their configured app and install automation, with the existing accepted-work contract.", Scopes: org, Targets: org},
	}
}
