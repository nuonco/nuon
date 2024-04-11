import { components } from '@/types/nuon-oapi-v3'

// component
export type TComponent = components['schemas']['app.Component']
export type TComponentConfig = {
  id: string
  component_id: string
  docker_build?: Record<string, unknown>
  external_image?: Record<string, unknown>
  helm?: Record<string, unknown>
  terraform_module?: Record<string, unknown>
  job?: Record<string, unknown>
}

// org
export type TOrg = components['schemas']['app.Org']

// install
export type TInstall = components['schemas']['app.Install']
export type TInstallAzureAccount = components['schemas']['app.AzureAccount']
export type TInstallAwsAccount = components['schemas']['app.AWSAccount']
export type TInstallComponent = components['schemas']['app.InstallComponent']
export type TInstallEvent = components['schemas']['app.InstallEvent']
export type TInstallDeploy = components['schemas']['app.InstallDeploy']

// sandbox
export type TSandbox = components['schemas']['app.Sandbox']
export type TSandboxConfig = components['schemas']['app.AppSandboxConfig']
export type TSandboxRun = components['schemas']['app.InstallSandboxRun']

// vcs configs
export type TVCSGitHub = components['schemas']['app.ConnectedGithubVCSConfig']
export type TVCSGit = components['schemas']['app.PublicGitVCSConfig']
