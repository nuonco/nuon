import { components } from '@/types/nuon-oapi-v3'

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
export type TBuild = {
  id: string
  created_at: string
  updated_at: string
  vcs_connection_commit: TVCSCommit
}

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
  }
export type TInstallEvent = Omit<
  components['schemas']['app.InstallEvent'],
  'payload'
> & {
  payload: string
}
export type TInstallDeploy = components['schemas']['app.InstallDeploy'] & {
  org_id: string
}

// sandbox
export type TSandbox = components['schemas']['app.Sandbox']
export type TSandboxConfig = components['schemas']['app.AppSandboxConfig'] & {
  cloud_platform?: string
}
export type TSandboxRun = components['schemas']['app.InstallSandboxRun']

// vcs configs
export type TVCSGitHub = components['schemas']['app.ConnectedGithubVCSConfig']
export type TVCSGit = components['schemas']['app.PublicGitVCSConfig']
export type TVCSCommit = components['schemas']['app.VCSConnectionCommit']
