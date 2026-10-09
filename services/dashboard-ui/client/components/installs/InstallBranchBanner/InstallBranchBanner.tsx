import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAPIError } from '@/types'

export interface IInstallBranchBanner {
  branchHref?: string
  branchName?: string
  commitsBehind?: number | null
  error?: TAPIError | null
  onRetry?: () => void
  onViewCommits?: () => void
}

export const InstallBranchBanner = ({
  branchHref,
  branchName,
  commitsBehind,
  error,
  onRetry,
  onViewCommits,
}: IInstallBranchBanner) => {
  if (error) {
    return (
      <Banner theme="error" className="!py-2" data-testid="branch-banner-error">
        <div className="flex flex-wrap justify-between items-center gap-3 w-full">
          <Text variant="subtext">
            Unable to load branch tracking for this install.
          </Text>
          {onRetry ? (
            <Button size="sm" onClick={onRetry}>
              Try again
            </Button>
          ) : null}
        </div>
      </Banner>
    )
  }

  if (!commitsBehind || commitsBehind < 1) return null

  const branch = (
    <Text as="span" family="mono" variant="subtext">
      {branchName ?? 'its app branch'}
    </Text>
  )

  return (
    <Banner
      theme="info"
      className="!py-2 dark:!text-blue-300"
      data-testid="branch-banner"
    >
      <div className="flex flex-wrap justify-between items-center gap-3 w-full">
        <Text variant="subtext">
          This install is{' '}
          <strong>
            {commitsBehind} {commitsBehind === 1 ? 'commit' : 'commits'} behind
          </strong>{' '}
          {branchHref ? <Link href={branchHref}>{branch}</Link> : branch}.
        </Text>
        {onViewCommits ? (
          <Button size="sm" onClick={onViewCommits}>
            View commits
          </Button>
        ) : null}
      </div>
    </Banner>
  )
}
