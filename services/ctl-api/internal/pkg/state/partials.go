package state

// PartialName identifies a component of install state that can be independently regenerated.
type PartialName string

const (
	PartialOrg        PartialName = "org"
	PartialApp        PartialName = "app"
	PartialDomain     PartialName = "domain"
	PartialRunner     PartialName = "runner"
	PartialCloud      PartialName = "cloud"
	PartialActions    PartialName = "actions"
	PartialInputs     PartialName = "inputs"
	PartialComponents PartialName = "components"
	PartialSandbox    PartialName = "sandbox"
	PartialStack      PartialName = "stack"
	PartialSecrets    PartialName = "secrets"
)

var AllPartials = []PartialName{
	PartialOrg,
	PartialApp,
	PartialDomain,
	PartialRunner,
	PartialCloud,
	PartialActions,
	PartialInputs,
	PartialComponents,
	PartialSandbox,
	PartialStack,
	PartialSecrets,
}

type PartialTarget struct {
	Name     PartialName
	EntityID string
}

type HintType string

const (
	HintDeployCompleted      HintType = "deploy-completed"
	HintComponentTeardown    HintType = "component-teardown"
	HintSandboxProvisioned   HintType = "sandbox-provisioned"
	HintSandboxDeprovisioned HintType = "sandbox-deprovisioned"
	HintSandboxReprovisioned HintType = "sandbox-reprovisioned"
	HintActionRan            HintType = "action-ran"
	HintStackRunCompleted    HintType = "stack-run-completed"
	HintStackOutputsUpdated  HintType = "stack-outputs-updated"
	HintInputsUpdated        HintType = "inputs-updated"
	HintSecretsUpdated       HintType = "secrets-updated"
	HintRunnerUpdated        HintType = "runner-updated"
	HintAppConfigUpdated     HintType = "app-config-updated"
	HintInstallCreated       HintType = "install-created"
)

var HintToPartials = map[HintType][]PartialName{
	HintDeployCompleted:      {PartialComponents},
	HintComponentTeardown:    {PartialComponents},
	HintSandboxProvisioned:   {PartialSandbox, PartialDomain},
	HintSandboxDeprovisioned: {PartialSandbox, PartialDomain},
	HintSandboxReprovisioned: {PartialSandbox, PartialDomain},
	HintActionRan:            {PartialActions},
	HintStackRunCompleted:    {PartialStack},
	HintStackOutputsUpdated:  {PartialStack, PartialInputs},
	HintInputsUpdated:        {PartialInputs},
	HintSecretsUpdated:       {PartialSecrets},
	HintRunnerUpdated:        {PartialRunner},
	HintAppConfigUpdated:     {PartialApp, PartialInputs, PartialActions},
	HintInstallCreated:       {PartialOrg, PartialApp, PartialRunner, PartialCloud, PartialInputs},
}

func TargetsForHint(hintType HintType, entityID string) []PartialTarget {
	partials := HintToPartials[hintType]
	targets := make([]PartialTarget, len(partials))
	for i, p := range partials {
		targets[i] = PartialTarget{Name: p, EntityID: entityID}
	}
	return targets
}

func AllPartialTargets() []PartialTarget {
	targets := make([]PartialTarget, len(AllPartials))
	for i, p := range AllPartials {
		targets[i] = PartialTarget{Name: p}
	}
	return targets
}
