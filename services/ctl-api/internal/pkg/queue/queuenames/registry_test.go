package queuenames

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryDefaults(t *testing.T) {
	tests := map[string]struct {
		ownerType string
		name      string
	}{
		"app branches": {ownerType: OwnerAppBranches, name: AppBranchDefaultQueueName},
		"apps":         {ownerType: OwnerApps, name: AppDefaultQueueName},
		"components":   {ownerType: OwnerComponents, name: ComponentDefaultQueueName},
		"general":      {ownerType: OwnerGeneral, name: GeneralSignalsQueueName},
		"installs":     {ownerType: OwnerInstalls, name: InstallSignalsQueueName},
		"onboardings":  {ownerType: OwnerOnboardings, name: OnboardingDefaultQueueName},
		"orgs":         {ownerType: OwnerOrgs, name: OrgSignalsQueueName},
		"runners":      {ownerType: OwnerRunners, name: RunnerSignalsQueueName},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			defaultName, ok := Default(test.ownerType)
			require.True(t, ok)
			require.Equal(t, test.name, defaultName)
			_, ok = SpecByName(test.ownerType, defaultName)
			require.True(t, ok)
		})
	}
}

func TestRegistrySoleOwners(t *testing.T) {
	tests := map[string]string{
		"notebooks":                 OwnerNotebooks,
		"vcs connections":           OwnerVCSConnections,
		"vcs webhook subscriptions": OwnerVCSWebhookSubscriptions,
	}
	for name, ownerType := range tests {
		t.Run(name, func(t *testing.T) {
			require.True(t, Sole(ownerType))
			_, hasDefault := Default(ownerType)
			require.False(t, hasDefault)
		})
	}
}

func TestFlowQueuesAreRegistered(t *testing.T) {
	for _, ownerType := range []string{OwnerAppBranches, OwnerApps, OwnerInstalls} {
		flow, ok := Flow(ownerType)
		require.True(t, ok)
		for _, name := range []string{
			flow.Workflows,
			flow.StepGroups,
			flow.Steps,
			flow.StepTargets,
			flow.GenerateSteps,
		} {
			_, ok := SpecByName(ownerType, name)
			require.Truef(t, ok, "%s queue %q is not registered", ownerType, name)
		}
	}
}

func TestInstallApprovalQueueIsBelowSignalTargets(t *testing.T) {
	_, ok := SpecByName(OwnerInstalls, InstallApprovalsQueueName)
	require.True(t, ok)

	flow, ok := Flow(OwnerInstalls)
	require.True(t, ok)
	require.NotEqual(t, flow.StepTargets, InstallApprovalsQueueName)
}

func TestRegistrySpecsAreValid(t *testing.T) {
	for ownerType, specs := range registry {
		t.Run(ownerType, func(t *testing.T) {
			names := make(map[string]struct{}, len(specs))
			for _, spec := range specs {
				_, duplicate := names[spec.Name]
				require.Falsef(t, duplicate, "duplicate queue name %q", spec.Name)
				names[spec.Name] = struct{}{}
				require.Positivef(t, spec.MaxDepth, "queue %q must have a positive max depth", spec.Name)
				if ownerType != OwnerRunners || spec.MaxInFlight != 0 {
					require.Positivef(t, spec.MaxInFlight, "queue %q must have a positive max in flight", spec.Name)
				}
			}
		})
	}
}
