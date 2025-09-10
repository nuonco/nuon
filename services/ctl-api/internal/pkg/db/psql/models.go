package psql

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// declare all models in the correct order they should be migrated.
func AllModels() []interface{} {
	return []interface{}{
		&app.EndpointAudit{},

		// management, auth and user management
		&app.Role{},
		&app.Account{},
		&app.AccountRole{},
		&app.Token{},
		&app.Policy{},

		&app.NotificationsConfig{},

		// org basics
		&app.Org{},
		&app.OrgInvite{},

		// installers
		&app.Installer{},
		&app.InstallerApp{},
		&app.InstallerMetadata{},

		// vcs basics
		&app.VCSConnection{},
		&app.VCSConnectionCommit{},

		// apps
		&app.App{},
		&app.AppRepository{},
		&app.AppConfig{},
		&app.AppBranch{},
		&app.AppSandboxConfig{},
		&app.AppRunnerConfig{},
		&app.AppInput{},
		&app.AppInputGroup{},
		&app.AppInputConfig{},
		&app.AppSecret{},
		&app.AppSecretsConfig{},
		&app.AppSecretConfig{},
		&app.AppPoliciesConfig{},
		&app.AppPolicyConfig{},
		&app.AppPermissionsConfig{},
		&app.AppAWSIAMRoleConfig{},
		&app.AppAWSIAMPolicyConfig{},
		&app.AppBreakGlassConfig{},
		&app.AppStackConfig{},

		// installs
		&app.AWSAccount{},
		&app.AzureAccount{},
		&app.Install{},
		&app.InstallState{},
		&app.InstallEvent{},
		&app.InstallInputs{},
		&app.InstallSandbox{},
		&app.InstallSandboxRun{},
		&app.InstallIntermediateData{},
		&app.InstallConfig{},
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
		&app.ExternalImageComponentConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.PublicGitVCSConfig{},
		&app.AWSECRImageConfig{},

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

		// actions
		&app.ActionWorkflow{},
		&app.ActionWorkflowConfig{},
		&app.ActionWorkflowStepConfig{},
		&app.ActionWorkflowTriggerConfig{},
		&app.InstallActionWorkflow{},
		&app.InstallActionWorkflowRun{},
		&app.InstallActionWorkflowManualTrigger{},
		&app.InstallActionWorkflowRunStep{},

		// install workflows
		&app.Workflow{},
		&app.WorkflowStep{},
		&app.WorkflowStepApproval{},
		&app.WorkflowStepApprovalResponse{},
		&app.WorkflowStepPolicyValidation{},

		// internal
		&migrations.MigrationModel{},

		// waitlist
		&app.Waitlist{},
		// NOTE(jm): this is a special table used in both ch and postgres
		&app.PSQLTableSize{},

		&app.TemporalPayload{},
	}
}
