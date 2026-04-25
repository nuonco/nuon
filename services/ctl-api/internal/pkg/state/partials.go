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

// AllPartials is the ordered list of all state partials.
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

// HintType describes what changed, allowing the workflow to determine which partials to regenerate.
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
)

// HintToPartials maps a hint type to the set of partials that should be regenerated.
var HintToPartials = map[HintType][]PartialName{
	HintDeployCompleted:      {PartialComponents},
	HintComponentTeardown:    {PartialComponents},
	HintSandboxProvisioned:   {PartialSandbox, PartialDomain},
	HintSandboxDeprovisioned: {PartialSandbox, PartialDomain},
	HintSandboxReprovisioned: {PartialSandbox, PartialDomain},
	HintActionRan:            {PartialActions},
	HintStackRunCompleted:    {PartialStack},
	HintStackOutputsUpdated:  {PartialStack},
	HintInputsUpdated:        {PartialInputs},
	HintSecretsUpdated:       {PartialSecrets},
	HintRunnerUpdated:        {PartialRunner},
	HintAppConfigUpdated:     {PartialApp, PartialInputs},
}

func allPartialsSet() map[PartialName]bool {
	m := make(map[PartialName]bool, len(AllPartials))
	for _, p := range AllPartials {
		m[p] = true
	}
	return m
}
