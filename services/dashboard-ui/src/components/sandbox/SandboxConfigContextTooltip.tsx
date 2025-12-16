'use client'

import cronstrue from 'cronstrue'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { SandboxEnvironmentVariablesModal, SandboxVariablesFilesModal } from './SandboxConfigModals'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TSandboxConfig, TVCSGit, TVCSGitHub, TCloudPlatform, TAppConfig } from '@/types'

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

function getSandboxConfigItems(
  config: TSandboxConfig,
  addModal: (modal: React.ReactNode) => string
): TContextTooltipItem[] {
  const items: TContextTooltipItem[] = []

  // Terraform version
  if (config?.terraform_version) {
    items.push({
      id: `config-terraform-version`,
      title: 'Terraform version',
      subtitle: config.terraform_version,
    })
  }

  // Cloud platform
  if (config?.cloud_platform) {
    items.push({
      id: `config-cloud-platform`,
      title: 'Cloud platform',
      subtitle: <CloudPlatform platform={config.cloud_platform as TCloudPlatform} />,
    })
  }

  // AWS region type
  if (config?.aws_region_type) {
    items.push({
      id: `config-aws-region-type`,
      title: 'AWS region type',
      subtitle: config.aws_region_type,
    })
  }

  // Drift schedule
  if (config?.drift_schedule) {
    let scheduleDescription = config.drift_schedule
    try {
      // Convert cron expression to human-readable text
      scheduleDescription = cronstrue.toString(config.drift_schedule, { 
        throwExceptionOnParseError: false,
        verbose: false 
      })
    } catch {
      // If parsing fails, fall back to the raw cron expression
      scheduleDescription = config.drift_schedule
    }

    items.push({
      id: `config-drift-schedule`,
      title: 'Drift schedule',
      subtitle: (
        <div className="flex flex-col gap-1">
          <Text variant="label" theme="neutral">{scheduleDescription}</Text>
          <Text variant="label" family="mono" theme="neutral">
            {config.drift_schedule}
          </Text>
        </div>
      ),
    })
  }

  // Environment variables
  if (config?.env_vars && Object.keys(config.env_vars).length > 0) {
    items.push({
      id: `config-env-vars`,
      title: 'Environment variables',
      leftContent: <Icon variant="List" />,
      onClick: () => {
        const modal = (
          <SandboxEnvironmentVariablesModal
            envVars={config.env_vars!}
          />
        )
        addModal(modal)
      },
      subtitle: 'View list',
    })
  }

  // References
  if (config?.references && config.references.length > 0) {
    items.push({
      id: `config-references`,
      title: 'References',
      subtitle: `${config.references.length} reference${config.references.length !== 1 ? 's' : ''}`,
    })
  }

  // VCS configuration
  const vcsConfig = config?.connected_github_vcs_config || config?.public_git_vcs_config
  if (vcsConfig) {
    items.push(...getConfigVCSItems(vcsConfig))
  }

  return items
}

interface SandboxConfigContextTooltipProps {
  appConfigId: string
  appId: string
  children?: React.ReactNode
}

export const SandboxConfigContextTooltip = ({
  appConfigId,
  appId,
  children,
}: SandboxConfigContextTooltipProps) => {
  const { org } = useOrg()
  const { addModal } = useSurfaces()

  const {
    data: appConfig,
    isLoading,
    error,
  } = useQuery<TAppConfig>({
    path: `/api/orgs/${org?.id}/apps/${appId}/configs/${appConfigId}?recurse=true`,
    enabled: !!org?.id && !!appId && !!appConfigId,
  })

  const config = appConfig?.sandbox

  if (isLoading) {
    return <Skeleton width="100px" height="20px" />
  }

  if (error || !config) {
    return null
  }

  return (
    <ContextTooltip
      title="Sandbox configuration"
      width="w-60"
      items={[
        {
          id: `config-app-config-id`,
          title: 'App Config ID',
          subtitle: <ID variant="label">{appConfigId}</ID>,
        },
        ...getSandboxConfigItems(config, addModal),
      ]}
    >
      {children || (
        <Text weight="strong" className="!flex items-center gap-1">
          <Link
            href={`/${org.id}/apps/${appId}`}
          >
            Sandbox configuration
          </Link>
          <Icon variant="Question" />
        </Text>
      )}
    </ContextTooltip>
  )
}
