'use client'

import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
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

function getConfigItems(config: TComponentConfig): TContextTooltipItem[] {
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
        ].filter(Boolean),
        ...getConfigVCSItems(
          config?.helm?.connected_github_vcs_config ||
            config?.helm?.public_git_vcs_config
        )
      )
      break
    case 'terraform_module':
      items.push(
        {
          id: `config-tf-version`,
          title: 'Terraform version',
          subtitle: config?.terraform_module?.version,
        },
        ...getConfigVCSItems(
          config?.terraform_module?.connected_github_vcs_config ||
            config?.terraform_module?.public_git_vcs_config
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

export const InstallComponentConfigCard = ({
  config,
}: {
  config: TComponentConfig
}) => {
  const { install } = useInstall()

  return (
    <ContextTooltip
      width="w-fit"
      title="Component configuration"
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
        ...getConfigItems(config),
      ]}
    >
      <Card className="!p-2 !flex-row">
        <Text weight="strong">
          <Link
            href={`/${install?.org_id}/apps/${install?.app_id}/configs/${install?.app_config_id}/components/${config?.component_id}`}
          >
            Component configuration <Icon variant="Question" />
          </Link>
        </Text>
      </Card>
    </ContextTooltip>
  )
}

export const InstallComponentConfigCardSkeleton = () => {
  return <Skeleton height="42px" width="200px" />
}
