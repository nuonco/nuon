import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Text } from '@/components/common/Text'
import type { TSandboxConfig, TVCSGit, TVCSGitHub } from '@/types'

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

interface ISandboxRunConfigCard {
  config: TSandboxConfig
  configHref: string
}

export const SandboxRunConfigCard = ({
  config,
  configHref,
}: ISandboxRunConfigCard) => {
  const items: TContextTooltipItem[] = []

  if (config?.drift_schedule) {
    items.push({
      id: `config-drift-schedule`,
      title: 'Drift schedule',
      subtitle: config.drift_schedule,
    })
  }

  if (config?.type === 'pulumi') {
    items.push({
      id: `config-version-`,
      title: 'Pulumi runtime',
      subtitle: config?.runtime ?? 'pulumi',
    })
  } else {
    items.push({
      id: `config-version-`,
      title: 'Terraform version',
      subtitle: config?.terraform_version,
    })
  }

  items.push(
    ...getConfigVCSItems(
      config?.connected_github_vcs_config || config?.public_git_vcs_config
    )
  )

  return (
    <ContextTooltip title="Component configuration" items={items}>
      <Card className="!p-2 !flex-row">
        <Text weight="strong">
          <Link href={configHref} variant="inline">
            Sanbox configuration <Icon variant="QuestionIcon" />
          </Link>
        </Text>
      </Card>
    </ContextTooltip>
  )
}
