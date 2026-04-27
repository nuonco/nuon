package labels

// Key, StepType, Domain, Name, and DisplayName are named string types so the
// compiler can distinguish label keys, value categories, and identifier slots
// in function signatures and struct fields. (Map-literal construction of
// Labels still requires string conversion since Labels remains map[string]string
// for JSONB compatibility.)
type (
	Key         string
	StepType    string
	Domain      string
	Name        string
	DisplayName string
)

// Common label keys.
const (
	KeyName          Key = "name"
	KeyDisplayName   Key = "display_name"
	KeyType          Key = "type"
	KeyDomain        Key = "domain"
	KeyComponentName Key = "component_name"
)

// Allowed values for the "type" label.
const (
	StepTypeSystem   StepType = "system"
	StepTypeUser     StepType = "user"
	StepTypeApproval StepType = "approval"
	StepTypeHidden   StepType = "hidden"
)

// Allowed values for the "domain" label.
const (
	DomainOther        Domain = "other"
	DomainInstallStack Domain = "install-stack"
	DomainSandbox      Domain = "sandbox"
	DomainAction       Domain = "action"
	DomainComponent    Domain = "component"
)

// Allowed values for the "name" label (machine identifiers).
const (
	NameGenerateInstallState          Name = "generate-install-state"
	NameProvisionRunnerServiceAccount Name = "provision-runner-service-account"
	NameProvisionInstallStack         Name = "provision-install-stack"
	NameAwaitRunnerHealth             Name = "await-runner-health"
	NameProvisionSandbox              Name = "provision-sandbox"
	NameSyncSecrets                   Name = "sync-secrets"
	NameProvisionSandboxDNS           Name = "provision-sandbox-dns"
	NameDeployComponentPrefix         Name = "deploy-"
)

// Allowed values for the "display_name" label (human-readable strings).
const (
	DisplayNameGenerateInstallState          DisplayName = "Generate install state"
	DisplayNameProvisionRunnerServiceAccount DisplayName = "Provision runner service account"
	DisplayNameProvisionInstallStack         DisplayName = "Provision install stack"
	DisplayNameAwaitRunnerHealth             DisplayName = "Await runner health"
	DisplayNameProvisionSandbox              DisplayName = "Provision sandbox"
	DisplayNameSyncSecrets                   DisplayName = "Sync secrets"
	DisplayNameProvisionSandboxDNS           DisplayName = "Provision sandbox DNS"
	DisplayNameDeployComponentPrefix         DisplayName = "Deploy "
)

// Identifier pairs a stable machine identifier with its human-readable display
// string. Centralized here so any consumer (workflow step groups, UI, etc.)
// references the same name/display pair.
type Identifier struct {
	Name        Name
	DisplayName DisplayName
}

var (
	GenerateInstallState = Identifier{
		Name:        NameGenerateInstallState,
		DisplayName: DisplayNameGenerateInstallState,
	}
	ProvisionRunnerServiceAccount = Identifier{
		Name:        NameProvisionRunnerServiceAccount,
		DisplayName: DisplayNameProvisionRunnerServiceAccount,
	}
	ProvisionInstallStack = Identifier{
		Name:        NameProvisionInstallStack,
		DisplayName: DisplayNameProvisionInstallStack,
	}
	AwaitRunnerHealth = Identifier{
		Name:        NameAwaitRunnerHealth,
		DisplayName: DisplayNameAwaitRunnerHealth,
	}
	ProvisionSandbox = Identifier{
		Name:        NameProvisionSandbox,
		DisplayName: DisplayNameProvisionSandbox,
	}
	SyncSecrets = Identifier{
		Name:        NameSyncSecrets,
		DisplayName: DisplayNameSyncSecrets,
	}
	ProvisionSandboxDNS = Identifier{
		Name:        NameProvisionSandboxDNS,
		DisplayName: DisplayNameProvisionSandboxDNS,
	}
	DeployComponentPrefix = Identifier{
		Name:        NameDeployComponentPrefix,
		DisplayName: DisplayNameDeployComponentPrefix,
	}
)
