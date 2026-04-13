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
import { PulumiConfigModal, PulumiEnvVarsModal } from '@/components/components/configs/PulumiConfig'
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

    case 'pulumi':
      items.push(
        {
          id: `config-pulumi-runtime`,
          title: 'Runtime',
          subtitle: config?.pulumi?.runtime,
        },
        {
          id: `config-pulumi-version`,
          title: 'Pulumi version',
          subtitle: config?.pulumi?.version,
        },
        ...[
          config?.pulumi?.config &&
            Object.keys(config.pulumi.config).length > 0 && {
              id: `config-pulumi-config-`,
              title: 'Config',
              leftContent: <Icon variant="List" />,
              onClick: () => {
                const modal = (
                  <PulumiConfigModal
                    config={config.pulumi!.config!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View list',
            },
          config?.pulumi?.env_vars &&
            Object.keys(config.pulumi.env_vars).length > 0 && {
              id: `config-pulumi-env-vars-`,
              title: 'Env vars',
              leftContent: <Icon variant="List" />,
              onClick: () => {
                const modal = (
                  <PulumiEnvVarsModal
                    envVars={config.pulumi!.env_vars!}
                  />
                )
                addModal(modal)
              },
              subtitle: 'View list',
            },
        ].filter(Boolean),
        ...getConfigVCSItems(
          config?.pulumi?.connected_github_vcs_config ||
            config?.pulumi?.public_git_vcs_config
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

interface IComponentConfigContextTooltip {
  config: TComponentConfig | null
  isLoading: boolean
  hasError: boolean
  orgId: string
  appId: string
  addModal: (modal: React.ReactNode) => string
  children?: React.ReactNode
}

export const ComponentConfigContextTooltip = ({
  config,
  isLoading,
  hasError,
  orgId,
  appId,
  addModal,
  children,
}: IComponentConfigContextTooltip) => {
  if (isLoading) {
    return <Skeleton width="100px" height="20px" />
  }

  if (hasError || !config) {
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
        <Text weight="strong" flex className="gap-1">
          <Link
            href={`/${orgId}/apps/${appId}/components/${config?.component_id}`}
          >
            Component configuration
          </Link>
          <Icon variant="Question" />
        </Text>
      )}
    </ContextTooltip>
  )
}
