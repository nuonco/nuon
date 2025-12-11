'use client'

import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { KubernetesManifestModal } from '@/components/components/configs/KubernetesConfig'
import { HelmValuesFilesModal, HelmValuesModal } from '@/components/components/configs/HelmConfig'
import { TerraformVariablesFilesModal, TerraformVariablesModal } from '@/components/components/configs/TerraformConfig'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TComponentConfig, TVCSGit, TVCSGitHub } from '@/types'

function getConfigVCSItems(
  vcsConfig: TVCSGit | TVCSGitHub
): TContextTooltipItem[] {
  return [
    vcsConfig?.repo && {
      id: `config-repo`,
      title: 'Repository',
      subtitle: vcsConfig?.repo,
    },
    vcsConfig?.directory && {
      id: `config-directory-`,
      title: 'Directory',
      subtitle: vcsConfig?.directory,
    },
    vcsConfig?.branch && {
      id: `config-branch-`,
      title: 'Branch',
      subtitle: vcsConfig?.branch,
    },
  ].filter(Boolean)
}

function getConfigItems(
  config: TComponentConfig,
  addModal: (modal: React.ReactNode) => string
): TContextTooltipItem[] {
  const items: TContextTooltipItem[] = []

  switch (config?.type) {
    case 'helm_chart':
      items.push(
        {
          id: `config-chart`,
          title: 'Chart name',
          subtitle: config?.helm?.chart_name,
        },
        ...[
          config?.helm?.namespace && {
            id: `config-namespace-`,
            title: 'Namespace',
            subtitle: config.helm.namespace,
          },
          config?.helm?.storage_driver && {
            id: `config-storage-driver-`,
            title: 'Storage driver',
            subtitle: config.helm.storage_driver,
          },
          config?.helm?.helm_config_json?.values_files &&
            config.helm.helm_config_json.values_files.length > 0 && {
              id: `config-values-files-`,
              title: 'Values files',
              leftContent: <Icon variant="FileCode" />,
              onClick: () => {
                const modal = (
                  <HelmValuesFilesModal
                    valuesFiles={config.helm.helm_config_json.values_files!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View YAML',
            },
          config?.helm?.helm_config_json?.values &&
            Object.keys(config.helm.helm_config_json.values).length > 0 && {
              id: `config-values-`,
              title: 'Values',
              leftContent: <Icon variant="List" />,
              onClick: () => {
                const modal = (
                  <HelmValuesModal
                    values={config.helm.helm_config_json.values!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View list',
            },
        ].filter(Boolean),
        ...getConfigVCSItems(
          config?.helm?.connected_github_vcs_config ||
            config?.helm?.public_git_vcs_config
        )
      )
      break

    case 'kubernetes_manifest':
      items.push(
        ...[
          config?.kubernetes_manifest?.namespace && {
            id: `config-namespace-`,
            title: 'Namespace',
            subtitle: config.kubernetes_manifest.namespace,
          },
          config?.kubernetes_manifest?.manifest && {
            id: `config-manifest-`,
            title: 'Manifest',
            leftContent: <Icon variant="FileCode" />,
            onClick: () => {
              const modal = (
                <KubernetesManifestModal
                  manifest={config.kubernetes_manifest.manifest!}
                />
              )
              addModal(modal)
            },
            subtitle: 'View YAML',
          },
        ].filter(Boolean)
      )
      break
    case 'terraform_module':
      items.push(
        {
          id: `config-tf-version`,
          title: 'Terraform version',
          subtitle: config?.terraform_module?.version,
        },
        ...[
          config?.terraform_module?.variables_files &&
            config.terraform_module.variables_files.length > 0 && {
              id: `config-variables-files-`,
              title: 'Variables files',
              leftContent: <Icon variant="FileCode" />,
              onClick: () => {
                const modal = (
                  <TerraformVariablesFilesModal
                    variablesFiles={config.terraform_module.variables_files!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View HCL',
            },
          config?.terraform_module?.variables &&
            Object.keys(config.terraform_module.variables).length > 0 && {
              id: `config-variables-`,
              title: 'Variables',
              leftContent: <Icon variant="List" />,
              onClick: () => {
                const modal = (
                  <TerraformVariablesModal
                    variables={config.terraform_module.variables!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View list',
            },
        ].filter(Boolean),
        ...getConfigVCSItems(
          config?.terraform_module?.connected_github_vcs_config ||
            config?.terraform_module?.public_git_vcs_config
        )
      )
      break

    case 'docker_build':
      items.push(
        ...[
          config?.docker_build?.dockerfile && {
            id: `config-dockerfile-`,
            title: 'Dockerfile',
            subtitle: config.docker_build.dockerfile,
          },
          config?.docker_build?.target && {
            id: `config-target-`,
            title: 'Target',
            subtitle: config.docker_build.target,
          },
        ].filter(Boolean),
        ...getConfigVCSItems(
          config?.docker_build?.connected_github_vcs_config ||
            config?.docker_build?.public_git_vcs_config
        )
      )
      break

    case 'external_image':
      items.push(
        {
          id: `config-url`,
          title: 'Image url',
          subtitle: config?.external_image?.image_url,
        },
        {
          id: `config-tag`,
          title: 'Image tag',
          subtitle: config?.external_image?.tag,
        }
      )
      break
    default:
  }

  return items
}

interface ComponentConfigContextTooltipProps {
  componentId: string
  configId: string
  appId: string
  children?: React.ReactNode
}

export const ComponentConfigContextTooltip = ({
  componentId,
  configId,
  appId,
  children,
}: ComponentConfigContextTooltipProps) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const {
    data: config,
    isLoading,
    error,
  } = useQuery<TComponentConfig>({
    path: `/api/orgs/${org?.id}/apps/${appId}/components/${componentId}/configs/${configId}`,
    enabled: !!org?.id && !!appId && !!componentId && !!configId,
  })

  if (isLoading) {
    return <Skeleton width="100px" height="20px" />
  }

  if (error || !config) {
    return null
  }

  return (
    <ContextTooltip
      title="Component configuration"
      width="w-60"
      items={[
        {
          id: `config-component-id-`,
          title: 'Component ID',
          subtitle: <ID variant="label">{config?.component_id}</ID>,
        },
        {
          id: `config-version-`,
          title: 'Config version',
          subtitle: config?.version?.toString(),
        },
        ...getConfigItems(config, addModal),
      ]}
    >
      {children || (
        <Text weight="strong" className="!flex items-center gap-1">
          <Link
            href={`/${org.id}/apps/${appId}/components/${config?.component_id}`}
          >
            Component configuration
          </Link>
          <Icon variant="Question" />
        </Text>
      )}
    </ContextTooltip>
  )
}
