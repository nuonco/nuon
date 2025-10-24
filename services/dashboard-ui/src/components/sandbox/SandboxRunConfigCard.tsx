'use client'

import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import {
  ContextTooltip,
  type TContextTooltipItem,
} from '@/components/common/ContextTooltip'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
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

export const SandboxRunConfigCard = ({
  config,
}: {
  config: TSandboxConfig
}) => {
  const { install } = useInstall()

  return (
    <ContextTooltip
      title="Component configuration"
      items={[
        {
          id: `config-version-`,
          title: 'Terraform version',
          subtitle: config?.terraform_version,
        },
        ...getConfigVCSItems(
          config?.connected_github_vcs_config || config?.public_git_vcs_config
        ),
      ]}
    >
      <Card className="!p-2 !flex-row">
        <Text weight="strong">
          <Link
            href={`/${install?.org_id}/apps/${install?.app_id}/configs/${install?.app_config_id}/sandbox`}
          >
            Sanbox configuration <Icon variant="Question" />
          </Link>
        </Text>
      </Card>
    </ContextTooltip>
  )
}
