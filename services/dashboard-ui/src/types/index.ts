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

// deploys
export type TInstallDeploy = components['schemas']['app.InstallDeploy'] & {
  org_id: string
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

// log stream
export type TLogStream = components['schemas']['app.LogStream']

// action workflows
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
