import { components } from '@/types/nuon-oapi-v3'

// app
export type TApp = components['schemas']['app.App']
export type TAppConfig = components['schemas']['app.AppConfig']
export type TAppInputConfig = components['schemas']['app.AppInputConfig']
export type TAppRunnerConfig = components['schemas']['app.AppRunnerConfig']
export type TAppSandboxConfig = components['schemas']['app.AppSandboxConfig']

// component
export type TComponent = components['schemas']['app.Component']
export type TComponentConfig =
  components['schemas']['app.ComponentConfigConnection']

// build
export type TComponentBuild = components['schemas']['app.ComponentBuild']
export type TBuild = TComponentBuild & { org_id: string }

// org
export type TOrg = components['schemas']['app.Org']

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
export type TInstallComponentSummary =
  components['schemas']['app.InstallComponentSummary']

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

// log stream
export type TLogStream = components['schemas']['app.LogStream']

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
export type TAction = components["schemas"]["app.ActionWorkflow"];;
export type TInstallActionRun =
  components["schemas"]["app.InstallActionWorkflowRun"];
export type TInstallAction =
  components["schemas"]["app.InstallActionWorkflow"];


// App / Install Readme
export type TReadme = components['schemas']['service.Readme']

// Waitlist
export type TWaitlist = components['schemas']['app.Waitlist']

// User / Account
export type TAccount = components['schemas']['app.Account']
export type TInvite = components['schemas']['app.OrgInvite']

// install workflows
export type TInstallWorkflow = components['schemas']['app.Workflow']
export type TInstallWorkflowStep =
  components['schemas']['app.WorkflowStep']
export type TWorkflow = components['schemas']['app.Workflow']
export type TWorkflowStep =
  components['schemas']['app.WorkflowStep']
export type TWorkflowStepApproval =
  components["schemas"]["app.WorkflowStepApproval"];
export type TWorkflowStepApprovalResponse =
  components["schemas"]["app.WorkflowStepApprovalResponse"];
export type TWorkflowStepApprovalType =
  components["schemas"]["app.WorkflowStepApprovalType"];

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
  git_ref: string
  version: string
}
