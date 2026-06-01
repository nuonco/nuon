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
import { SandboxEnvironmentVariablesModal, SandboxVariablesFilesModal } from '../SandboxConfigModals'
import type { TSandboxConfig, TVCSGit, TVCSGitHub, TCloudPlatform } from '@/types'

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
  const isPulumi = config?.type === 'pulumi'

  if (isPulumi) {
    items.push({
      id: `config-type`,
      title: 'Type',
      subtitle: 'Pulumi',
    })

    if (config?.runtime) {
      items.push({
        id: `config-runtime`,
        title: 'Runtime',
        subtitle: config.runtime,
      })
    }

    if (config?.pulumi_version) {
      items.push({
        id: `config-pulumi-version`,
        title: 'Pulumi version',
        subtitle: config.pulumi_version,
      })
    }
  }

  if (!isPulumi && config?.terraform_version) {
    items.push({
      id: `config-terraform-version`,
      title: 'Terraform version',
      subtitle: config.terraform_version,
    })
  }

  if (config?.cloud_platform) {
    items.push({
      id: `config-cloud-platform`,
      title: 'Cloud platform',
      subtitle: <CloudPlatform platform={config.cloud_platform as TCloudPlatform} />,
    })
  }

  if (config?.aws_region_type) {
    items.push({
      id: `config-aws-region-type`,
      title: 'AWS region type',
      subtitle: config.aws_region_type,
    })
  }

  if (config?.drift_schedule) {
    let scheduleDescription = config.drift_schedule
    try {
      scheduleDescription = cronstrue.toString(config.drift_schedule, {
        throwExceptionOnParseError: false,
        verbose: false,
      })
    } catch {
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

  if (config?.env_vars && Object.keys(config.env_vars).length > 0) {
    items.push({
      id: `config-env-vars`,
      title: 'Environment variables',
      leftContent: <Icon variant="ListIcon" />,
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

  if (config?.references && config.references.length > 0) {
    items.push({
      id: `config-references`,
      title: 'References',
      subtitle: `${config.references.length} reference${config.references.length !== 1 ? 's' : ''}`,
    })
  }

  const vcsConfig = config?.connected_github_vcs_config || config?.public_git_vcs_config
  if (vcsConfig) {
    items.push(...getConfigVCSItems(vcsConfig))
  }

  return items
}

interface ISandboxConfigContextTooltip {
  appConfigId: string
  orgId: string
  appId: string
  config: TSandboxConfig | undefined
  isLoading: boolean
  error: any
  addModal: (modal: React.ReactNode) => string
  children?: React.ReactNode
}

export const SandboxConfigContextTooltip = ({
  appConfigId,
  orgId,
  appId,
  config,
  isLoading,
  error,
  addModal,
  children,
}: ISandboxConfigContextTooltip) => {
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
        <Text weight="strong" flex className="gap-1">
          <Link
            href={`/${orgId}/apps/${appId}`}
          >
            Sandbox configuration
          </Link>
          <Icon variant="QuestionIcon" />
        </Text>
      )}
    </ContextTooltip>
  )
}
