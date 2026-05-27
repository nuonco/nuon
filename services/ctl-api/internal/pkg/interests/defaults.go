package interests

// AllEvents returns the new-subscription default: a sentinel Interests config
// that matches every supported lifecycle and approval event. This is what
// the "Send all events" toggle in the picker writes when it's checked, and
// what new webhooks/slack subs default to.
func AllEvents() Interests {
	return Interests{AllEvents: true}
}

// Default returns the materialised "power user opted out of AllEvents"
// baseline. Five resources (installs, stacks, components, sandboxes,
// install_configurations) are present with empty ops (= all sub-ops),
// Outcome=Completion, and approval flags both true (where supported).
// Runners + actions are absent (= off).
//
// Toggling "Send all events" off in the picker writes this exact shape so
// users land on a sensible per-resource starting point instead of an empty
// config that silently drops every event.
func Default() Interests {
	return Interests{
		Resources: map[ResourceKind]ResourceCfg{
			ResourceInstalls: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
			},
			ResourceStacks: {
				Outcome: OutcomeCompletion,
			},
			ResourceComponents: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
				DriftDetected:     true,
			},
			ResourceSandboxes: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
				DriftDetected:     true,
			},
			ResourceInstallConfigurations: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
			},
		},
	}
}
