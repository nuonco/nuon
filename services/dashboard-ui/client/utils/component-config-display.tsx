import type { TComponentConfig, TVCSGit, TVCSGitHub } from '@/types'

export interface IConfigDisplayItem {
  label: string
  value: string | number | null | undefined
  type?: 'text' | 'code' | 'link' | 'cron'
}

export interface IConfigVCSInfo {
  repo?: string
  directory?: string
  branch?: string
  vcsConfig?: TVCSGit | TVCSGitHub
}

export interface IConfigDisplayData {
  commonFields: {
    version?: number
    type?: string
    buildTimeout?: string
    deployTimeout?: string
    driftSchedule?: string
    checksum?: string
  }
  typeSpecificFields: IConfigDisplayItem[]
  vcsInfo: IConfigVCSInfo | null
  operationRoles?: Record<string, string> | null
}

function formatTimeout(timeout?: string): string | undefined {
  if (!timeout) return undefined
  return timeout
}

function getVCSInfo(
  vcsConfig?: TVCSGit | TVCSGitHub
): IConfigVCSInfo | null {
  if (!vcsConfig) return null

  return {
    repo: vcsConfig.repo,
    directory: vcsConfig.directory,
    branch: vcsConfig.branch,
    vcsConfig,
  }
}

export function getComponentConfigDisplayData(
  config: TComponentConfig
): IConfigDisplayData {
  const commonFields = {
    version: config.version,
    type: config.type,
    buildTimeout: formatTimeout(config.build_timeout),
    deployTimeout: formatTimeout(config.deploy_timeout),
    driftSchedule: config.drift_schedule,
    checksum: config.checksum,
  }

  let typeSpecificFields: IConfigDisplayItem[] = []
  let vcsInfo: IConfigVCSInfo | null = null

  switch (config.type) {
    case 'helm_chart':
      typeSpecificFields = [
        {
          label: 'Chart name',
          value: config.helm?.chart_name,
        },
        {
          label: 'Namespace',
          value: config.helm?.namespace,
        },
        {
          label: 'Storage driver',
          value: config.helm?.storage_driver,
        },
      ]
      vcsInfo = getVCSInfo(
        config.helm?.connected_github_vcs_config ||
          config.helm?.public_git_vcs_config
      )
      break

    case 'terraform_module':
      typeSpecificFields = [
        {
          label: 'Terraform version',
          value: config.terraform_module?.version,
        },
      ]
      vcsInfo = getVCSInfo(
        config.terraform_module?.connected_github_vcs_config ||
          config.terraform_module?.public_git_vcs_config
      )
      break

    case 'kubernetes_manifest':
      typeSpecificFields = [
        {
          label: 'Namespace',
          value: config.kubernetes_manifest?.namespace,
        },
      ]
      vcsInfo = getVCSInfo(
        config.kubernetes_manifest?.connected_github_vcs_config ||
          config.kubernetes_manifest?.public_git_vcs_config
      )
      break

    case 'docker_build':
      typeSpecificFields = [
        {
          label: 'Dockerfile name',
          value: config.docker_build?.dockerfile,
        },
        {
          label: 'Target',
          value: config.docker_build?.target,
        },
      ]
      vcsInfo = getVCSInfo(
        config.docker_build?.connected_github_vcs_config ||
          config.docker_build?.public_git_vcs_config
      )
      break

    case 'external_image':
      typeSpecificFields = [
        {
          label: 'Image URL',
          value: config.external_image?.image_url,
        },
        {
          label: 'Image tag',
          value: config.external_image?.tag,
        },
      ]
      break

    case 'job':
      typeSpecificFields = [
        {
          label: 'Image URL',
          value: config.job?.image_url,
        },
        {
          label: 'Tag',
          value: config.job?.tag,
        },
        {
          label: 'Command',
          value: config.job?.cmd?.join(' '),
        },
        {
          label: 'Arguments',
          value: config.job?.args?.join(' '),
        },
      ]
      break

    case 'unknown':
    default:
      typeSpecificFields = [
        {
          label: 'Type',
          value: config.type || 'Unknown component type',
        },
      ]
      break
  }

  return {
    commonFields,
    typeSpecificFields: typeSpecificFields.filter((field) => field.value),
    vcsInfo,
    operationRoles: config.operation_roles,
  }
}
