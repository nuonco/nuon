import type { TAppBranch } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Badge } from '../../atoms/Badge'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import { CommitSummary } from '../../molecules/CommitSummary'
import { OverviewCard, OverviewCardGrid } from '../../molecules/OverviewCard'

export interface IAppBranchOverviewCards {
  branch?: TAppBranch
  installCount?: number
  hasMoreInstalls?: boolean
  isLoading?: boolean
}

export const AppBranchOverviewCards = ({
  branch,
  installCount,
  hasMoreInstalls = false,
  isLoading = false,
}: IAppBranchOverviewCards) => {
  const config = branch ? latestBranchConfig(branch) : undefined
  const vcs =
    config?.connected_github_vcs_config ?? config?.public_git_vcs_config
  const repo =
    config?.connected_github_vcs_config?.repo ??
    config?.public_git_vcs_config?.repo
  const commit = branch?.latest_run?.vcs_connection_commit

  return (
    <OverviewCardGrid columns={3}>
      <OverviewCard title="Branch info">
        <span className="flex items-center gap-2">
          <Icon variant="GitBranchIcon" size={18} />
          <Text
            family="mono"
            weight="medium"
            loading={isLoading}
            loadingWidth={12}
          >
            {branch?.name ?? '—'}
          </Text>
        </span>
        <span className="flex flex-wrap items-center gap-2">
          {isLoading ? (
            <Badge loading loadingWidth={8} />
          ) : config?.config_number ? (
            <Badge>Config v{config.config_number}</Badge>
          ) : null}
          <Text
            variant="caption"
            family="mono"
            color="tertiary"
            loading={isLoading}
            loadingWidth={18}
            lines={1}
          >
            {repo ?? vcs?.directory ?? 'No repository configured'}
          </Text>
        </span>
      </OverviewCard>

      <OverviewCard title="Last update">
        <CommitSummary
          commit={commit}
          updatedAt={branch?.latest_run?.updated_at}
          isLoading={isLoading}
        />
      </OverviewCard>

      <OverviewCard title="Installs">
        <Text
          variant="title"
          loading={isLoading || installCount === undefined}
          loadingWidth={3}
        >
          {hasMoreInstalls ? `${installCount}+` : installCount}
        </Text>
        <Text variant="caption" color="tertiary">
          Assigned to this branch
        </Text>
      </OverviewCard>
    </OverviewCardGrid>
  )
}
