import { components } from '@/types/nuon-oapi-v3'

// logs
export type TWaypointLogTerminalEvent = {
  line?: { msg: string }
  raw?: { data: string }
  step?: { msg: string }
  status?: { msg: string }
}

export type TWaypointLog = {
  Complete: Record<string, unknown>
  Open: Record<string, unknown>
  State: Record<string, unknown>
  Terminal: {
    buffered: boolean
    events: Array<TWaypointLogTerminalEvent>
  }
}

// component
export type TComponent = components['schemas']['app.Component']
export type TComponentConfig = {
  id: string
  component_id: string
  docker_build?: Record<string, any>
  external_image?: Record<string, any>
  helm?: Record<string, any>
  terraform_module?: Record<string, any>
  job?: Record<string, any>
}

// build
export type TComponentBuild = components['schemas']['app.ComponentBuild']
export type TBuild = TComponentBuild & { org_id: string }
export type TComponentBuildLogs = Array<TWaypointLog>
export type TComponentBuildPlan = Record<string, any>

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
export type TInstallDeployLogs = Array<TWaypointLog>
export type TInstallDeployPlan = Record<string, any>

// sandbox
export type TSandbox = components['schemas']['app.Sandbox']
export type TSandboxConfig = components['schemas']['app.AppSandboxConfig'] & {
  cloud_platform?: string
}
export type TSandboxRun = components['schemas']['app.InstallSandboxRun'] & {
  org_id: string
}
export type TSandboxRunLogs = Array<TWaypointLog>

// vcs configs
export type TVCSGitHub = components['schemas']['app.ConnectedGithubVCSConfig']
export type TVCSGit = components['schemas']['app.PublicGitVCSConfig']
export type TVCSCommit = components['schemas']['app.VCSConnectionCommit']
