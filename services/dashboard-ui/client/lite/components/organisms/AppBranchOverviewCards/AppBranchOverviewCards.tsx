import type { TAppBranch } from '@/types'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Badge } from '../../atoms/Badge'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import { Link } from '../../atoms/Link'
import { CommitSummary } from '../../molecules/CommitSummary'
import { OverviewCard, OverviewCardGrid } from '../../molecules/OverviewCard'

export interface IAppBranchOverviewCards {
  branch?: TAppBranch
  installCount?: number
  hasMoreInstalls?: boolean
  loading?: boolean
}

export const AppBranchOverviewCards = ({
  branch,
  installCount,
  hasMoreInstalls = false,
  loading = false,
}: IAppBranchOverviewCards) => {
  const config = branch ? latestBranchConfig(branch) : undefined
  const vcs =
    config?.connected_github_vcs_config ?? config?.public_git_vcs_config
  const repo =
    config?.connected_github_vcs_config?.repo ??
    config?.public_git_vcs_config?.repo
  const repoHref = repo
    ? repo.startsWith('http')
      ? repo
      : `https://github.com/${repo}`
    : undefined
  const commit = branch?.latest_run?.vcs_connection_commit

  return (
    <OverviewCardGrid columns={3}>
      <OverviewCard title="Branch">
        <span className="flex items-center gap-2">
          <Icon variant="GitBranchIcon" size={18} />
          <Text
            family="mono"
            weight="medium"
            loading={loading}
            loadingWidth={12}
          >
            {branch?.name ?? '—'}
          </Text>
        </span>
        <span className="flex min-w-0 items-center gap-2">
          <Icon variant="GitHub" size={16} />
          {loading ? (
            <Text variant="caption" family="mono" loading loadingWidth={18} />
          ) : repoHref ? (
            <Link
              href={repoHref}
              external
              variant="caption"
              className="min-w-0 truncate font-mono"
            >
              {repo}
            </Link>
          ) : (
            <Text variant="caption" family="mono" color="tertiary" lines={1}>
              {vcs?.directory ?? 'No repository configured'}
            </Text>
          )}
        </span>
        <span className="flex flex-wrap items-center gap-2">
          {loading ? (
            <Badge loading loadingWidth={8} />
          ) : config?.config_number ? (
            <Badge>Config v{config.config_number}</Badge>
          ) : null}
        </span>
      </OverviewCard>

      <OverviewCard title="Last update">
        <CommitSummary
          commit={commit}
          updatedAt={branch?.latest_run?.updated_at}
          loading={loading}
        />
      </OverviewCard>

      <OverviewCard title="Installs">
        <Text
          variant="title"
          loading={loading || installCount === undefined}
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
