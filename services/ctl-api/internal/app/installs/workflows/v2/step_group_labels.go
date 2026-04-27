package v2

import (
	"strings"

	"github.com/nuonco/nuon/pkg/labels"
)

// Each constructor returns a fresh labels.Labels so callers cannot mutate a
// shared instance.

func stepGroupLabels(id labels.Identifier, typ labels.StepType, domain labels.Domain) labels.Labels {
	return labels.Labels{
		string(labels.KeyName):        string(id.Name),
		string(labels.KeyDisplayName): string(id.DisplayName),
		string(labels.KeyType):        string(typ),
		string(labels.KeyDomain):      string(domain),
	}
}

func StepGroupGenerateInstallState() labels.Labels {
	return stepGroupLabels(labels.GenerateInstallState, labels.StepTypeHidden, labels.DomainOther)
}

func StepGroupProvisionRunnerServiceAccount() labels.Labels {
	return stepGroupLabels(labels.ProvisionRunnerServiceAccount, labels.StepTypeSystem, labels.DomainInstallStack)
}

func StepGroupProvisionInstallStack() labels.Labels {
	return stepGroupLabels(labels.ProvisionInstallStack, labels.StepTypeUser, labels.DomainInstallStack)
}

func StepGroupAwaitRunnerHealth() labels.Labels {
	return stepGroupLabels(labels.AwaitRunnerHealth, labels.StepTypeSystem, labels.DomainOther)
}

func StepGroupProvisionSandbox() labels.Labels {
	return stepGroupLabels(labels.ProvisionSandbox, labels.StepTypeApproval, labels.DomainSandbox)
}

func StepGroupSyncSecrets() labels.Labels {
	return stepGroupLabels(labels.SyncSecrets, labels.StepTypeSystem, labels.DomainSandbox)
}

func StepGroupProvisionSandboxDNS() labels.Labels {
	return stepGroupLabels(labels.ProvisionSandboxDNS, labels.StepTypeSystem, labels.DomainSandbox)
}

func StepGroupDeployComponent(name string, isImage bool) labels.Labels {
	typ := labels.StepTypeApproval
	if isImage {
		typ = labels.StepTypeSystem
	}
	return labels.Labels{
		string(labels.KeyName):          string(labels.DeployComponentPrefix.Name) + name,
		string(labels.KeyDisplayName):   string(labels.DeployComponentPrefix.DisplayName) + name,
		string(labels.KeyType):          string(typ),
		string(labels.KeyDomain):        string(labels.DomainComponent),
		string(labels.KeyComponentName): name,
	}
}

func StepGroupActionTrigger(triggerName string) labels.Labels {
	display := triggerName
	if len(triggerName) > 0 {
		display = strings.ToUpper(triggerName[:1]) + triggerName[1:]
	}
	return labels.Labels{
		string(labels.KeyName):        triggerName,
		string(labels.KeyDisplayName): display,
		string(labels.KeyType):        string(labels.StepTypeSystem),
		string(labels.KeyDomain):      string(labels.DomainAction),
	}
}
