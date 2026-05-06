import { components } from '@/types/nuon-oapi-v3'
import type { Interests as TInterests } from '@/components/interests/types'

// app branches
export type TAppBranch = components['schemas']['app.AppBranch']
export type TAppBranchConfig = components['schemas']['app.AppBranchConfig']
export type TAppBranchInstallGroup = components['schemas']['app.AppBranchInstallGroup']
export type TAppBranchRun = components['schemas']['app.AppBranchRun']
export type TCreateAppBranchRequest = components['schemas']['service.CreateAppBranchRequest']
export type TVCSBranch = { name: string }

// app
export type TApp = components['schemas']['app.App']
export type TAppConfig = components['schemas']['app.AppConfig']
export type TAppInputConfig = components['schemas']['app.AppInputConfig']
export type TAppRunnerConfig = components['schemas']['app.AppRunnerConfig']
export type TAppSandboxConfig = components['schemas']['app.AppSandboxConfig']
export type TAppSandboxBuild = {
  id: string
  created_by_id?: string
  created_by?: { email?: string }
  created_at?: string
  updated_at?: string
  org_id?: string
  app_id?: string
  app_config_id?: string
  app_sandbox_config_id?: string
  status?: string
  status_description?: string
  status_v2?: { status?: string; status_human_description?: string; metadata?: { [key: string]: unknown } }
  log_stream?: { id?: string; open?: boolean }
  runner_job?: { id: string }
  vcs_connection_commit?: { sha?: string; message?: string }
}
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

// policy reports
export type TPolicyReport = components['schemas']['app.PolicyReport']
export type TPolicyReportOwnerType =
  components['schemas']['app.PolicyReportOwnerType']
export type TPolicyResult = components['schemas']['app.PolicyResult']
export type TPolicyViolation = components['schemas']['app.PolicyViolation']
export type TPolicyInputRef = components['schemas']['app.PolicyInputRef']

// policy analytics
export type TPolicyAnalyticsSummary =
  components['schemas']['service.PolicyAnalyticsSummary']
export type TPolicyAnalyticsTimeseries =
  components['schemas']['service.PolicyAnalyticsTimeseries']
export type TTimeseriesBucket =
  components['schemas']['service.TimeseriesBucket']
export type TSeriesPoint = components['schemas']['service.SeriesPoint']
export type TPolicyAnalyticsBreakdown =
  components['schemas']['service.PolicyAnalyticsBreakdown']
export type TBreakdownEntry = components['schemas']['service.BreakdownEntry']

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

// webhooks
//
// `interests` is stamped `swaggertype:"object"` on the Go side, so the
// auto-generated SDK shape is a generic object. Re-cast to the hand-written
// Interests type from @/components/interests at the API boundary instead.
export type TWebhook = Omit<
  components['schemas']['service.CurrentOrgWebhookResponse'],
  'interests'
> & {
  interests?: TInterests
}
export type TCreateWebhookBody = Omit<
  components['schemas']['service.CreateCurrentOrgWebhookRequest'],
  'interests'
> & {
  interests?: TInterests
}

// install
export type TInstall = components['schemas']['app.Install'] & {
  app?: components['schemas']['app.App']
  created_by?: components['schemas']['app.Account']
  gcp_account?: { project_id?: string; region?: string }
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
export type TInstallRoles = components['schemas']['app.InstallRoles']
export type TInstallRoleUsage = components['schemas']['app.InstallRoleUsage']
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
    app: { id: string; secrets: Record<string, string> }
    components: Record<
      string,
      {
        outputs: Record<string, string>
      }
    >
    install: {
      internal_domain: string
      public_domain: string
      inputs: Record<string, string>
      sandbox: {
        outputs: {
          account: {
            id: string
            region: string
          }
          cluster: {
            arn: string
            certificate_authority_data: string
            cluster_security_group_id: string
            endpoint: string
            name: string
            node_security_group_id: string
            oidc_issuer_url: string
            platform_version: string
            status: string
          }
          ecr: {
            registry_id: string
            registry_url: string
            repository_arn: string
            repository_name: string
            repository_url: string
          }
          internal_domain: {
            name: string
            nameservers: string[]
            zone_id: string
          }
          public_domain: {
            name: string
            nameservers: string[]
            zone_id: string
          }
          runner: {
            odr_iam_role_arn: string
            runner_iam_role_arn: string
          }
          vpc: {
            azs: string[]
            cidr: string
            default_security_group_id: string
            id: string
            name: string
            private_subnet_cidr_blocks: string[]
            private_subnet_ids: string[]
            public_subnet_cidr_blocks: string[]
            public_subnet_ids: string[]
          }
        }
      }
    }
  }
}
export type TInstallDeployPlan = {
  actual: {
    waypoint_plan: {
      waypoint_job: {
        hcl_config: string
      }
      variables: {
        intermediate_data: TInstallDeployPlanIntermediateData
      }
    }
  }
}

// sandbox
export type TSandboxConfig = components['schemas']['app.AppSandboxConfig'] & {
  cloud_platform?: string
}
export type TSandboxRun = components['schemas']['app.InstallSandboxRun'] & {
  org_id: string
}

// vcs configs
export type TVCSConnection = components['schemas']['app.VCSConnection']
export type TVCSGitHub = components['schemas']['app.ConnectedGithubVCSConfig']
export type TVCSGit = components['schemas']['app.PublicGitVCSConfig']
export type TVCSCommit = components['schemas']['app.VCSConnectionCommit']
export type TVCSConnectionStatus = {
  status: 'active' | 'suspended' | 'unknown'
  github_install_id: string
  account: {
    login: string
    id: number
    type: string
  } | null
  suspended_at: string | null
  suspended_by: {
    login: string
    id: number
  } | null
  permissions: Record<string, string>
  repository_selection: 'all' | 'selected'
  checked_at: string
  error?: string
}
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

// OTEL logs
export type TOTELLog = components['schemas']['app.OtelLogRecord']

// runner
export type TRunnerGroup = components['schemas']['app.RunnerGroup']
export type TRunnerGroupSettings =
  components['schemas']['app.RunnerGroupSettings']
export type TRunnerGroupType = components['schemas']['app.RunnerGroupType']
export type TRunner = components['schemas']['app.Runner']
export type TRunnerJob = components['schemas']['app.RunnerJob']
export type TRunnerHealthCheck = components['schemas']['app.RunnerHealthCheck']
export type TRunnerHeartbeat = components['schemas']['app.RunnerHeartBeat']
export type TRunnerMngHeartbeat = {
  build: TRunnerHeartbeat
  install: TRunnerHeartbeat
  mng: TRunnerHeartbeat
  org: TRunnerHeartbeat
}
export type TRunnerSettings = components['schemas']['app.RunnerGroupSettings']
export type TRunnerJobPlan = Record<string, any>
export type TRunnerProcess = components['schemas']['app.RunnerProcess'] & {
  labels?: string[]
}
export type TRunnerProcessShutdown =
  components['schemas']['app.RunnerProcessShutdown']

// log stream
export type TLogStream = components['schemas']['app.LogStream']

// otel spans (Phase 4 spans endpoint — defined manually until ctl-api ships
// the schema; the wire shape mirrors the runner -> ctl-api ingestion record).
export type TSpan = {
  span_id: string
  parent_span_id?: string
  trace_id?: string
  name: string
  start_time: string
  end_time: string
  duration_ns: number
  status_code?: 'Ok' | 'Error' | 'Unset' | string
  status_message?: string
  attributes?: Record<string, string>
  scope_name?: string
  service_name?: string
}

// old action workflows types
export type TActionWorkflow = components['schemas']['app.ActionWorkflow']
export type TActionConfig = components['schemas']['app.ActionWorkflowConfig']
export type TActionConfigStep =
  components['schemas']['app.ActionWorkflowStepConfig']
export type TActionConfigTrigger =
  components['schemas']['app.ActionWorkflowTriggerConfig']
export type TActionConfigTriggerType =
  components['schemas']['app.ActionWorkflowTriggerType']
export type TInstallActionWorkflowRun =
  components['schemas']['app.InstallActionWorkflowRun']
export type TInstallActionWorkflow =
  components['schemas']['app.InstallActionWorkflow']

// new action types
export type TAction = components['schemas']['app.ActionWorkflow']
export type TInstallActionRun =
  components['schemas']['app.InstallActionWorkflowRun']
export type TInstallAction = components['schemas']['app.InstallActionWorkflow']

// App / Install Readme
export type TReadme = components['schemas']['service.Readme']

// Waitlist
export type TWaitlist = components['schemas']['app.Waitlist']

// User / Account
export type TAccount = components['schemas']['app.Account']
export type TInvite = components['schemas']['app.OrgInvite']

// User Journey (Enhanced with completion tracking and metadata)
export interface TUserJourneyStep {
  name: string
  title: string
  complete: boolean
  completed_at: string | null
  completion_method: 'auto' | 'manual' | 'cli' | 'api' | null
  completion_source: 'dashboard' | 'cli' | 'api' | 'system' | null
  metadata: Record<string, any>
}

export interface TUserJourney {
  name: string
  title: string
  steps: TUserJourneyStep[]
}

// install workflows
export type TInstallWorkflow = components['schemas']['app.Workflow']
export type TInstallWorkflowStep = components['schemas']['app.WorkflowStep']
export type TWorkflow = components['schemas']['app.Workflow']
export type TWorkflowStep = components['schemas']['app.WorkflowStep'] & {
  log_stream?: { id?: string; open?: boolean }
}
export type TWorkflowStepApproval =
  components['schemas']['app.WorkflowStepApproval']
export type TWorkflowStepApprovalResponse = { type: string } & any
export type TWorkflowStepApprovalType =
  components['schemas']['app.WorkflowStepApprovalType']
export type TWorkflowResponse = components['schemas']['app.WorkflowResponse']

// app secrets
export type TAppSecretConfig = components['schemas']['app.AppSecretConfig']
export type TAppSecretsConfig = components['schemas']['app.AppSecretsConfig']

// app / install stack
export type TInstallStack = components['schemas']['app.InstallStack']
export type TInstallStackVersion =
  components['schemas']['app.InstallStackVersion']
export type TInstallStackVersionRun =
  components['schemas']['app.InstallStackVersionRun']
export type TInstallStackOutputs =
  components['schemas']['app.InstallStackOutputs']
export type TAppStackConfig = components['schemas']['app.AppStackConfig']

// api version
export type TAPIVersion = {
  ui: { version: string; git_ref: string }
  api: { version: string; git_ref: string }
}

// terraform workspaces
export type TTerraformWorkspaceState =
  components['schemas']['app.TerraformWorkspaceStateJSON']
export type TTerraformWorkspaceLock =
  components['schemas']['app.TerraformWorkspaceLock']
export type TTerraformState = {
  format_version: string
  terraform_version: string
  values: {
    outputs?: {
      [key: string]: {
        sensitive: boolean
        value: any
        type: string
      }
    }
    root_module?: {
      resources?: Array<{
        address: string
        mode: string
        type: string
        name: string
        provider_name: string
        schema_version: number
        values: Record<string, any>
        sensitive_values: Record<string, any>
      }>
      child_modules?: Array<{
        resources?: Array<{
          address: string
          mode: string
          type: string
          name: string
          provider_name: string
          schema_version: number
          index?: number | string
          values: Record<string, any>
          sensitive_values: Record<string, any>
          depends_on?: string[]
        }>
        address?: string
      }>
    }
  }
}

// available roles
export type TAvailableRole = components['schemas']['service.AvailableRole']
export type TAvailableRolesResponse =
  components['schemas']['service.AvailableRolesResponse']
export type TOperationType = components['schemas']['app.OperationType']
export type TPrincipalType = 'component' | 'sandbox' | 'action'

// composite status
export type TCompositeStatus = components['schemas']['app.CompositeStatus']

// onboarding
export type TOnboarding = components['schemas']['app.Onboarding']
export type TExampleApp = components['schemas']['service.ExampleApp']
export type TCompleteOrganizationStepRequest =
  components['schemas']['service.CompleteOrganizationStepRequest']
export type TCompleteYourStackStepRequest =
  components['schemas']['service.CompleteYourStackStepRequest']
export type TCompleteInstallStepRequest =
  components['schemas']['service.CompleteInstallStepRequest']

// sandbox admin types (not swagger-annotated, defined manually)
export type TAdminSandboxConfig = {
  id: string
  runner_id: string
  job_type: string
  created_at: string
  updated_at: string
  preset: string
  duration: number
  fault_rate: number
  error_message: string
  fail_at_step: string
  sleep_duration: number
  timeout: number
  trigger_shutdown: boolean
  log_lines: string[] | null
  plan_contents: string
  outputs: any
}

export type TSandboxRunner = {
  id: string
  runner_group_id: string
  created_at: string
  updated_at: string
  status: string
  status_description: string
  connected: boolean
  runner_group?: {
    org_id: string
    org?: {
      name: string
    }
  }
}

export type TSandboxLogTemplate = {
  key: string
  name: string
  lines: string[]
}

export type TSandboxPlanTemplate = {
  key: string
  name: string
  contents: string
}

export type TSandboxTemplates = {
  log_templates: TSandboxLogTemplate[]
  plan_templates: TSandboxPlanTemplate[]
}

export type TSandboxJob = {
  id: string
  runner_id: string
  job_type: string
  status: string
  status_description: string
  created_at: string
  updated_at: string
}

// auth
export type TMe = {
  id: string
  email: string
  identities: Array<{
    picture?: string
    name?: string
  }>
  [key: string]: any
}
// TODO(nnnnat): use the generated type once it is ready
// components['schemas']['service.AuthMeResponse']
