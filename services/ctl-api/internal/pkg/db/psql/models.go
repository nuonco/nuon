package psql

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// declare all models in the correct order they should be migrated.
func AllModels() []any {
	return []any{
		&app.EndpointAudit{},

		// management, auth and user management
		&app.Role{},
		&app.Account{},
		&app.AccountRole{},
		&app.Token{},
		&app.Policy{},
		&app.IdentityProvider{},
		&app.AccountIdentity{},
		&app.DeviceCode{},
		&app.OIDCTrustPolicy{},

		&app.NotificationsConfig{},

		// org basics
		&app.Org{},
		&app.Webhook{},
		&app.OrgInvite{},
		&app.AWSAccountConnection{},

		// slack
		&app.SlackInstallation{},
		&app.SlackOrgLink{},
		&app.SlackChannelSubscription{},
		&app.SlackThreadAnchor{},

		// installers
		&app.Installer{},
		&app.InstallerApp{},
		&app.InstallRoles{},
		&app.InstallRoleUsage{},
		&app.InstallerMetadata{},

		// vcs basics
		&app.VCSConnection{},
		&app.VCSConnectionCommit{},
		&app.VCSWebhookSubscription{},
		&app.GithubEvent{},
		&app.VCSConnectionEvent{},

		// apps
		&app.App{},
		&app.AppRepository{},
		&app.AppConfig{},
		&app.AppBranch{},
		&app.AppBranchConfig{},
		&app.AppBranchInstallGroup{},
		&app.AppBranchRun{},
		&app.Trigger{},
		&app.TriggerSecret{},
		&app.TriggerEvent{},
		&app.TriggerRule{},
		&app.EventDispatch{},
		&app.EventRunbookWaiter{},
		&app.InstallGroupRun{},
		&app.InstallAppConfigVersion{},
		&app.InstallConfigSync{},
		&app.InstallConfigVersion{},
		&app.AppInstallsConfig{},
		&app.AppInstallConfigSync{},
		&app.InstallCreationApproval{},
		&app.AppSandboxConfig{},
		&app.AppSandboxBuild{},
		&app.AppRunnerConfig{},
		&app.AppInput{},
		&app.AppInputGroup{},
		&app.AppInputConfig{},
		&app.AppSecret{},
		&app.AppSecretsConfig{},
		&app.AppSecretConfig{},
		&app.AppSecretKubernetesSyncTarget{},
		&app.AppPoliciesConfig{},
		&app.AppPolicyConfig{},
		&app.AppPermissionsConfig{},
		&app.AppAWSIAMRoleConfig{},
		&app.AppAWSIAMPolicyConfig{},
		&app.AppBreakGlassConfig{},
		&app.AppStackConfig{},
		&app.AppOperationRoleConfig{},
		&app.AppOperationRoleRule{},
		&app.AppKubernetesContextsConfig{},
		&app.AppKubernetesContextConfig{},

		// installs
		&app.AWSAccount{},
		&app.AzureAccount{},
		&app.GCPAccount{},
		&app.Install{},
		&app.InstallState{},
		&app.InstallEvent{},
		&app.InstallInputs{},
		&app.InstallSandbox{},
		&app.InstallSandboxRun{},
		&app.InstallIntermediateData{},
		&app.InstallConfig{},
		&app.InstallAppBranchConnection{},
		&app.InstallAuditLog{},

		// install stacks
		&app.InstallStack{},
		&app.InstallStackOutputs{},
		&app.InstallStackVersion{},
		&app.InstallStackVersionRun{},

		// component configuration
		&app.Component{},
		&app.ComponentDependency{},
		&app.ComponentConfigConnection{},
		&app.HelmComponentConfig{},
		&app.TerraformModuleComponentConfig{},
		&app.DockerBuildComponentConfig{},
		&app.JobComponentConfig{},
		&app.KubernetesManifestComponentConfig{},
		&app.PulumiComponentConfig{},
		&app.ExternalImageComponentConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.PublicGitVCSConfig{},
		&app.AWSECRImageConfig{},
		&app.GCPGARImageConfig{},
		&app.AzureACRImageConfig{},

		// component management
		&app.ComponentBuild{},
		&app.ComponentRelease{},
		&app.ComponentReleaseStep{},

		// install management
		&app.InstallDeploy{},
		&app.InstallComponent{},

		// log streams
		&app.LogStream{},

		// runner jobs and groups
		&app.RunnerGroup{},
		&app.RunnerOperation{},
		&app.RunnerGroupSettings{},
		&app.Runner{},
		&app.RunnerProcess{},
		&app.RunnerProcessShutdown{},
		&app.RunnerJob{},
		&app.RunnerJobPlan{},
		&app.RunnerJobExecution{},
		&app.RunnerJobExecutionOutputs{},
		&app.RunnerJobExecutionResult{},
		&app.TerraformWorkspaceState{},
		&app.TerraformWorkspace{},
		&app.TerraformWorkspaceLock{},
		&app.TerraformWorkspaceStateJSON{},
		&app.OCIArtifact{},
		&app.HelmRelease{},
		&app.SandboxModeJobConfig{},
		&app.SandboxModeSignalConfig{},

		// queues
		&app.Queue{},
		&app.QueueSignal{},

		// actions
		&app.ActionWorkflow{},
		&app.ActionWorkflowConfig{},
		&app.ActionWorkflowStepConfig{},
		&app.ActionWorkflowTriggerConfig{},
		&app.InstallActionWorkflow{},
		&app.InstallActionWorkflowRun{},
		&app.InstallActionWorkflowManualTrigger{},
		&app.InstallActionWorkflowRunStep{},

		// notebooks
		&app.Notebook{},
		&app.NotebookCell{},
		&app.NotebookCellRun{},

		// runbooks
		&app.Runbook{},
		&app.RunbookConfig{},
		&app.RunbookStepConfig{},
		&app.RunbookInput{},
		&app.InstallRunbook{},
		&app.InstallRunbookRun{},

		// install workflows
		&app.Workflow{},
		&app.WorkflowStepGroup{},
		&app.WorkflowStep{},
		&app.WorkflowStepApproval{},
		&app.WorkflowStepApprovalResponse{},
		&app.WorkflowStepPolicyValidation{},
		&app.WorkflowRun{},
		&app.PolicyReport{},

		// internal
		&migrations.MigrationModel{},

		// drifts
		&app.DriftedObject{},

		// app label keys (query-only view)
		&app.AppLabelKey{},

		// waitlist
		&app.Waitlist{},
		// NOTE(jm): this is a special table used in both ch and postgres
		&app.PSQLTableSize{},

		&app.TemporalPayload{},
		&app.TemporalBlob{},

		// onboarding
		&app.Onboarding{},
	}
}
