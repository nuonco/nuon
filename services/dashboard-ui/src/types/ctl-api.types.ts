import { components } from '@/types/nuon-oapi-v3'

// app
export type TApp = components['schemas']['app.App']
export type TAppConfig = components['schemas']['app.AppConfig']
export type TAppInputConfig = components['schemas']['app.AppInputConfig']
export type TAppRunnerConfig = components['schemas']['app.AppRunnerConfig']
export type TAppSandboxConfig = components['schemas']['app.AppSandboxConfig']
//
// Policy types - manually defined as API schema may not be deployed yet
export type TAppPolicyType =
  | 'kubernetes_cluster'
  | 'terraform_module'
  | 'helm_chart'
  | 'kubernetes_manifest'
  | 'docker_build'
  | 'container_image'
  | 'sandbox'

export type TAppPolicyEngine = 'kyverno' | 'opa'

export type TAppPolicyConfig = {
  id?: string
  created_by_id?: string
  created_at?: string
  updated_at?: string
  org_id?: string
  app_id?: string
  app_config_id?: string
  app_policies_config?: string
  type?: TAppPolicyType
  engine?: TAppPolicyEngine
  name?: string
  contents?: string
  components?: string[]
}

export type TAppPoliciesConfig = {
  id?: string
  created_by_id?: string
  created_at?: string
  updated_at?: string
  org_id?: string
  app_id?: string
  app_config_id?: string
  policies?: TAppPolicyConfig[]
}
export type TAppBranch = components['schemas']['app.AppBranch']
export type TAppBranchConfig = components['schemas']['app.AppBranchConfig']
export type TAppBranchInstallGroup = components['schemas']['app.AppBranchInstallGroup']
export type TCreateAppBranchRequest =
  components['schemas']['service.CreateAppBranchRequest']

// policy reports
export type TPolicyReport = components['schemas']['app.PolicyReport']
export type TPolicyReportOwnerType =
  components['schemas']['app.PolicyReportOwnerType']
export type TPolicyResult = components['schemas']['app.PolicyResult']
export type TPolicyViolation = components['schemas']['app.PolicyViolation']
export type TPolicyInputRef = components['schemas']['app.PolicyInputRef']

// component
export type TComponent = components['schemas']['app.Component']
export type TComponentConfig =
  components['schemas']['app.ComponentConfigConnection']
export type TComponentType = components['schemas']['app.ComponentType']

// build
export type TComponentBuild = components['schemas']['app.ComponentBuild']
export type TBuild = TComponentBuild & { org_id: string }

// org
export type TOrg = components['schemas']['app.Org']
export type TOrgInvite = components['schemas']['app.OrgInvite']
export type TOrgStats = {
  install_names: string[]
  app_count: number
  install_count: number
}

// install
export type TInstall = components['schemas']['app.Install'] & {
  app?: components['schemas']['app.App']
  org_id?: string
}
export type TInstallAzureAccount = components['schemas']['app.AzureAccount']
export type TInstallAwsAccount = components['schemas']['app.AWSAccount']
export type TInstallComponent =
  components['schemas']['app.InstallComponent'] & {
    org_id?: string
    install_deploys?: Array<TInstallDeploy>
  }
export type TInstallEvent = Omit<
  components['schemas']['app.InstallEvent'],
  'payload'
> & {
  payload: string
}
export type TInstallInputs = components['schemas']['app.InstallInputs']
export type TInstallComponentOutputs = Record<string, string>
export type TInstallConfig = components['schemas']['app.InstallConfig']
export type TInstallAuditLog = components['schemas']['app.InstallAuditLog']
export type TDriftedObject = components['schemas']['app.DriftedObject']
// deploys
export type TInstallDeploy = components['schemas']['app.InstallDeploy'] & {
  org_id: string
}
export type TDeploy = TInstallDeploy
export type TInstallDeployPlanIntermediateData = {
  nuon: {
    helm?: {
      diff: {
        manifest: string
      }
    }
    kubernetes?: {
      diff: {
        manifest: string
      }
    }
    terraform?: {
      plan: string
    }
  }
}
export type TInstallStack = components['schemas']['app.InstallStack']

// action
export type TActionWorkflow = components['schemas']['app.ActionWorkflow']
export type TActionWorkflowConfig =
  components['schemas']['app.ActionWorkflowConfig']
export type TActionWorkflowStepConfig =
  components['schemas']['app.ActionWorkflowStepConfig']
export type TActionWorkflowTriggerConfig =
  components['schemas']['app.ActionWorkflowTriggerConfig']
export type TActionWorkflowTriggerType =
  components['schemas']['app.ActionWorkflowTriggerType']
export type TInstallActionWorkflowRun =
  components['schemas']['app.InstallActionWorkflowRun']
export type TInstallActionWorkflowStepRun =
  components['schemas']['app.InstallActionWorkflowStepRun']
export type TInstallActionWorkflowStepRunSummary = {
  action_workflow_id: string
  action_workflow_name: string
  action_workflow_run_id: string
  app_id: string
  install_id: string
  created_at: string
  updated_at: string
  status: string
  status_description: string
}

// workflows
export type TWorkflow = components['schemas']['app.Workflow']
export type TWorkflowEvent = components['schemas']['app.WorkflowEvent']
// Note: InstallWorkflow was renamed to Workflow in the backend
export type TInstallWorkflow = TWorkflow
export type TInstallWorkflowTask = components['schemas']['app.Task']
export type TInstallWorkflowDeployTask =
  components['schemas']['app.InstallWorkflowDeployTask']
export type TInstallWorkflowTaskAttempt =
  components['schemas']['app.TaskAttempt']

export type TRunnerJob = components['schemas']['app.RunnerJob']
export type TLogStream = components['schemas']['app.LogStream']
export type TStatus = components['schemas']['app.Status']
export type TCompositeStatus = components['schemas']['app.CompositeStatus']

// Health
export type TAPIVersionData = components['schemas']['public.Version']

// releases
export type TComponentRelease = components['schemas']['app.ComponentRelease']
export type TComponentReleaseStep =
  components['schemas']['app.ComponentReleaseStep']

// vcs
export type TVCSConnection = components['schemas']['app.VCSConnection']
export type TVCSConnectionCommit =
  components['schemas']['app.VCSConnectionCommit']
export type TConnectedGithubVCSConfig =
  components['schemas']['app.ConnectedGithubVCSConfig']
export type TPublicGitVCSConfig = components['schemas']['app.PublicGitVCSConfig']

// VCS Repositories
export type TVCSConnectionRepo = {
  id: number
  name: string
  full_name: string
  description?: string
  private: boolean
  fork: boolean
  html_url: string
  default_branch: string
  updated_at: string
}

export type TVCSConnectionReposResponse = {
  repositories: TVCSConnectionRepo[]
  total_count: number
}

// account
export type TAccount = components['schemas']['app.Account']
export type TRole = components['schemas']['app.Role']
export type TUserJourney = components['schemas']['app.UserJourney']
export type TUserJourneyStep = components['schemas']['app.UserJourneyStep']

// sandbox
export type TInstallSandbox = components['schemas']['app.InstallSandbox']
export type TInstallSandboxRun = components['schemas']['app.InstallSandboxRun']
