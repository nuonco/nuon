package interests

func AllEvents() Interests {
	return Interests{AllEvents: true}
}

func Default() Interests {
	return Interests{
		Resources: map[ResourceKind]ResourceCfg{
			ResourceInstalls: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
				InstallDegraded:   true,
			},
			ResourceStacks: {
				Outcome:       OutcomeCompletion,
				RoleChanges:   true,
				InputsUpdated: true,
			},
			ResourceComponents: {
				Outcome:           OutcomeCompletion,
				ApprovalRequests:  true,
				ApprovalResponses: true,
				DriftDetected:     true,
				ComponentHealth:   true,
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
			ResourceAppBranches: {
				Outcome:      OutcomeCompletion,
				ConfigSynced: true,
			},
		},
	}
}
