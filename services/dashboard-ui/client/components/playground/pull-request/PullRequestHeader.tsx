import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { ReactNode } from 'react'
import type { TPlaygroundPullRequest, TPullRequestState } from './fixtures'

const STATE_THEME: Record<TPullRequestState, 'success' | 'brand' | 'neutral'> =
  {
    open: 'success',
    merged: 'brand',
    closed: 'neutral',
  }

export interface IPullRequestHeader {
  pullRequest?: TPlaygroundPullRequest
  actions?: ReactNode
  backLink?: ReactNode
  loading?: boolean
}

export const PullRequestHeader = ({
  pullRequest,
  actions,
  backLink,
  loading = false,
}: IPullRequestHeader) => {
  const repo = pullRequest
    ? `${pullRequest.repo_owner}/${pullRequest.repo_name}`
    : undefined

  return (
    <div className="flex flex-col gap-2">
      {backLink}
      <SectionHeader
      loading={loading}
      title={
        pullRequest ? (
          <span className="flex items-baseline gap-2">
            <Text variant="h3" weight="stronger" theme="neutral" family="mono">
              #{pullRequest.number}
            </Text>
            <Text variant="h3" weight="stronger">
              {pullRequest.title}
            </Text>
          </span>
        ) : null
      }
      status={
        loading ? (
          <Badge size="sm" loading />
        ) : pullRequest ? (
          <span className="flex items-center gap-2">
            <Badge size="sm" theme={STATE_THEME[pullRequest.state]}>
              {pullRequest.state}
            </Badge>
            {pullRequest.is_draft ? (
              <Badge size="sm" theme="neutral">
                draft
              </Badge>
            ) : null}
          </span>
        ) : null
      }
      description={
        loading ? (
          <span className="flex items-center gap-2">
            <Text variant="subtext" loading loadingWidth={16} />
            <Text variant="subtext" loading loadingWidth={12} />
            <Text variant="subtext" loading loadingWidth={9} />
          </span>
        ) : pullRequest ? (
          <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
            {pullRequest.html_url ? (
              <Link
                href={pullRequest.html_url}
                isExternal
                isATag
                variant="inline"
                className="flex items-center gap-1"
              >
                <Icon variant="GitHub" size={12} />
                {repo}
              </Link>
            ) : (
              <span className="flex items-center gap-1">
                <Icon variant="GitHub" size={12} theme="neutral" />
                <Text variant="subtext" theme="neutral">
                  {repo}
                </Text>
              </span>
            )}
            <span aria-hidden>·</span>
            <span className="flex items-center gap-1">
              <Icon variant="GitBranchIcon" size={12} theme="neutral" />
              <Text variant="subtext" theme="neutral" family="mono">
                {pullRequest.head_ref}
              </Text>
              <Icon variant="ArrowRightIcon" size={11} theme="neutral" />
              <Text variant="subtext" theme="neutral" family="mono">
                {pullRequest.base_ref}
              </Text>
            </span>
            <span aria-hidden>·</span>
            <span className="flex items-center gap-1">
              <Icon variant="GitCommitIcon" size={12} theme="neutral" />
              <Text variant="subtext" theme="neutral" family="mono">
                {pullRequest.head_sha.slice(0, 7)}
              </Text>
            </span>
            {pullRequest.author_login ? (
              <>
                <span aria-hidden>·</span>
                <Text variant="subtext" theme="neutral">
                  opened by {pullRequest.author_login}
                </Text>
              </>
            ) : null}
            {pullRequest.opened_at ? (
              <Time
                variant="subtext"
                theme="neutral"
                time={pullRequest.opened_at}
                format="relative"
              />
            ) : null}
          </span>
        ) : null
      }
      actions={loading ? null : actions}
    />
    </div>
  )
}
