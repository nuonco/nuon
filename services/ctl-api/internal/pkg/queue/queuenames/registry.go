package queuenames

type Spec struct {
	Name        string
	MaxInFlight int
	MaxDepth    int
}

type FlowSpec struct {
	Workflows     string
	StepGroups    string
	Steps         string
	StepTargets   string
	GenerateSteps string
}

const (
	OwnerAppBranches             = "app_branches"
	OwnerApps                    = "apps"
	OwnerComponents              = "components"
	OwnerGeneral                 = "general"
	OwnerInstalls                = "installs"
	OwnerNotebooks               = "notebooks"
	OwnerOnboardings             = "onboardings"
	OwnerOrgs                    = "orgs"
	OwnerRunners                 = "runners"
	OwnerVCSConnections          = "vcs_connections"
	OwnerVCSWebhookSubscriptions = "vcs_webhook_subscriptions"
)

const (
	AppBranchDefaultQueueName            = ""
	AppBranchWorkflowsQueueName          = "app-branch-workflows"
	AppBranchGenerateStepsQueueName      = "app-branch-generate-steps"
	AppBranchWorkflowStepGroupsQueueName = "app-branch-workflow-step-groups"
	AppBranchWorkflowStepsQueueName      = "app-branch-workflow-steps"
	AppBranchSignalsQueueName            = "app-branch-signals"
	AppBranchSandboxBuildsQueueName      = "app-branch-sandbox-builds"

	AppDefaultQueueName            = ""
	AppTriggersQueueName           = "app-triggers"
	AppWorkflowsQueueName          = "app-workflows"
	AppGenerateStepsQueueName      = "app-generate-steps"
	AppWorkflowStepGroupsQueueName = "app-workflow-step-groups"
	AppWorkflowStepsQueueName      = "app-workflow-steps"
	AppSignalsQueueName            = "app-signals"
	AppInstallSyncsQueueName       = "app-install-syncs"

	ComponentDefaultQueueName       = ""
	ComponentWorkflowStepsQueueName = "component-workflow-steps"

	GeneralSignalsQueueName = "general-signals"

	InstallWorkflowsQueueName          = "install-workflows"
	InstallGenerateStepsQueueName      = "install-generate-steps"
	InstallWorkflowStepGroupsQueueName = "install-workflow-step-groups"
	InstallWorkflowStepsQueueName      = "install-workflow-steps"
	InstallSignalsQueueName            = "install-signals"
	InstallApprovalsQueueName          = "install-approvals"
	InstallStateManagerQueueName       = "state-manager"
	InstallActionWorkflowsQueueName    = "install-action-workflows"
	InstallDriftWorkflowsQueueName     = "install-drift-workflows"
	InstallActionCronSignalsQueueName  = "install-action-cron-signals"
	InstallDriftCronSignalsQueueName   = "install-drift-cron-signals"
	InstallComponentHealthQueueName    = "install-component-health"

	OnboardingDefaultQueueName   = ""
	OrgSignalsQueueName          = "org-signals"
	OrgHealthcheckCronsQueueName = "org-healthcheck-crons"

	RunnerSignalsQueueName          = "runner-signals"
	RunnerHealthcheckCronsQueueName = "runner-healthcheck-crons"
	RunnerHealthChecksQueueName     = "health-checks"
	RunnerSyncQueueName             = "sync"
	RunnerBuildQueueName            = "build"
	RunnerDeployQueueName           = "deploy"
	RunnerSandboxQueueName          = "sandbox"
	RunnerJobsQueueName             = "runner"
	RunnerOperationsQueueName       = "operations"
	RunnerManagementQueueName       = "management"
	RunnerActionsQueueName          = "actions"
	RunnerImageActionsQueueName     = "image-actions"
)

var registry = map[string][]Spec{
	OwnerAppBranches: {
		{Name: AppBranchDefaultQueueName, MaxInFlight: 25, MaxDepth: 50},
		{Name: AppBranchWorkflowsQueueName, MaxInFlight: 25, MaxDepth: 50},
		{Name: AppBranchGenerateStepsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: AppBranchWorkflowStepGroupsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: AppBranchWorkflowStepsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: AppBranchSignalsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: AppBranchSandboxBuildsQueueName, MaxInFlight: 2, MaxDepth: 50},
	},
	OwnerApps: {
		{Name: AppDefaultQueueName, MaxInFlight: 1, MaxDepth: 50},
		{Name: AppTriggersQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: AppWorkflowsQueueName, MaxInFlight: 2, MaxDepth: 50},
		{Name: AppGenerateStepsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: AppWorkflowStepGroupsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: AppWorkflowStepsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: AppSignalsQueueName, MaxInFlight: 20, MaxDepth: 50},
		{Name: AppInstallSyncsQueueName, MaxInFlight: 5, MaxDepth: 5},
	},
	OwnerComponents: {
		{Name: ComponentDefaultQueueName, MaxInFlight: 1, MaxDepth: 50},
		{Name: ComponentWorkflowStepsQueueName, MaxInFlight: 10, MaxDepth: 50},
	},
	OwnerGeneral: {
		{Name: GeneralSignalsQueueName, MaxInFlight: 5, MaxDepth: 50},
	},
	OwnerInstalls: {
		{Name: InstallWorkflowsQueueName, MaxInFlight: 25, MaxDepth: 50},
		{Name: InstallGenerateStepsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: InstallWorkflowStepGroupsQueueName, MaxInFlight: 40, MaxDepth: 50},
		{Name: InstallWorkflowStepsQueueName, MaxInFlight: 40, MaxDepth: 50},
		{Name: InstallSignalsQueueName, MaxInFlight: 20, MaxDepth: 50},
		{Name: InstallApprovalsQueueName, MaxInFlight: 20, MaxDepth: 50},
		{Name: InstallStateManagerQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: InstallActionWorkflowsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: InstallDriftWorkflowsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: InstallActionCronSignalsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: InstallDriftCronSignalsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: InstallComponentHealthQueueName, MaxInFlight: 1, MaxDepth: 10},
	},
	OwnerOnboardings: {
		{Name: OnboardingDefaultQueueName, MaxInFlight: 1, MaxDepth: 10},
	},
	OwnerOrgs: {
		{Name: OrgSignalsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: OrgHealthcheckCronsQueueName, MaxInFlight: 2, MaxDepth: 10},
	},
	OwnerRunners: {
		{Name: RunnerSignalsQueueName, MaxInFlight: 10, MaxDepth: 50},
		{Name: RunnerHealthcheckCronsQueueName, MaxInFlight: 5, MaxDepth: 50},
		{Name: RunnerHealthChecksQueueName, MaxDepth: 100},
		{Name: RunnerSyncQueueName, MaxDepth: 100},
		{Name: RunnerBuildQueueName, MaxDepth: 100},
		{Name: RunnerDeployQueueName, MaxDepth: 100},
		{Name: RunnerSandboxQueueName, MaxDepth: 100},
		{Name: RunnerJobsQueueName, MaxDepth: 100},
		{Name: RunnerOperationsQueueName, MaxDepth: 100},
		{Name: RunnerManagementQueueName, MaxDepth: 100},
		{Name: RunnerActionsQueueName, MaxDepth: 100},
		{Name: RunnerImageActionsQueueName, MaxDepth: 100},
	},
}

var defaults = map[string]string{
	OwnerAppBranches: AppBranchDefaultQueueName,
	OwnerApps:        AppDefaultQueueName,
	OwnerComponents:  ComponentDefaultQueueName,
	OwnerGeneral:     GeneralSignalsQueueName,
	OwnerInstalls:    InstallSignalsQueueName,
	OwnerOnboardings: OnboardingDefaultQueueName,
	OwnerOrgs:        OrgSignalsQueueName,
	OwnerRunners:     RunnerSignalsQueueName,
}

var soleOwnerTypes = map[string]struct{}{
	OwnerNotebooks:               {},
	OwnerVCSConnections:          {},
	OwnerVCSWebhookSubscriptions: {},
}

var flowRegistry = map[string]FlowSpec{
	OwnerAppBranches: {
		Workflows:     AppBranchWorkflowsQueueName,
		StepGroups:    AppBranchWorkflowStepGroupsQueueName,
		Steps:         AppBranchWorkflowStepsQueueName,
		StepTargets:   AppBranchSignalsQueueName,
		GenerateSteps: AppBranchGenerateStepsQueueName,
	},
	OwnerApps: {
		Workflows:     AppWorkflowsQueueName,
		StepGroups:    AppWorkflowStepGroupsQueueName,
		Steps:         AppWorkflowStepsQueueName,
		StepTargets:   AppSignalsQueueName,
		GenerateSteps: AppGenerateStepsQueueName,
	},
	OwnerInstalls: {
		Workflows:     InstallWorkflowsQueueName,
		StepGroups:    InstallWorkflowStepGroupsQueueName,
		Steps:         InstallWorkflowStepsQueueName,
		StepTargets:   InstallSignalsQueueName,
		GenerateSteps: InstallGenerateStepsQueueName,
	},
}

func Specs(ownerType string) ([]Spec, bool) {
	specs, ok := registry[ownerType]
	return specs, ok
}

func SpecByName(ownerType, name string) (Spec, bool) {
	specs, ok := Specs(ownerType)
	if !ok {
		return Spec{}, false
	}
	for _, spec := range specs {
		if spec.Name == name {
			return spec, true
		}
	}
	return Spec{}, false
}

func Default(ownerType string) (string, bool) {
	name, ok := defaults[ownerType]
	return name, ok
}

func Sole(ownerType string) bool {
	_, ok := soleOwnerTypes[ownerType]
	return ok
}

func Flow(ownerType string) (FlowSpec, bool) {
	spec, ok := flowRegistry[ownerType]
	return spec, ok
}

func RunnerJobGroupSpecs() []Spec {
	specs, _ := Specs(OwnerRunners)
	jobGroups := make([]Spec, 0, len(specs))
	for _, spec := range specs {
		if spec.MaxInFlight == 0 {
			jobGroups = append(jobGroups, spec)
		}
	}
	return jobGroups
}
